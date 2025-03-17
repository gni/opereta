package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"opereta/pkg/core"
	"opereta/pkg/inventory"
	"opereta/pkg/modules"
	"opereta/pkg/tasks"
)

const (
	defaultMaxRetries = 3
	defaultRetryDelay = 2 * time.Second
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[0;31m"
	ColorGreen  = "\033[0;32m"
	ColorYellow = "\033[0;33m"
	ColorBlue   = "\033[0;34m"
)

func printColored(color, message string) {
	fmt.Println(color + message + ColorReset)
}

func coloredMarker(color string) string {
	return color + "[-]" + ColorReset
}

// formatLaunchingTask prints the starting info for a task.
func formatLaunchingTask(hostGlobalShort, server, task, eventShort string) string {
	marker := coloredMarker(ColorGreen)
	return fmt.Sprintf("[%s][%s] %s %s task %s '%s'\n",
		hostGlobalShort, eventShort, marker, server+" "+marker, marker, task)
}

// formatSuccess now omits the redundant event id (it’s already in the header).
func formatSuccess(hostGlobalShort, server, task, result, eventShort, executedAt, duration string) string {
	marker := coloredMarker(ColorGreen)
	return fmt.Sprintf(
		"[%s][%s] %s %s %s Success %s\nExecuted At: %s | Duration: %s\nResult:\n%s\n",
		hostGlobalShort, eventShort, marker, server, marker, task, executedAt, duration, result)
}

// Similarly, formatError no longer repeats the event id.
func formatError(hostGlobalShort, server, message, eventShort, executedAt, duration string) string {

	errorMarker := coloredMarker(ColorRed)
	return fmt.Sprintf(
		"[%s][%s] %s %s %s Error %s\nExecuted At: %s | Duration: %s\n",
		hostGlobalShort, eventShort, errorMarker, server, errorMarker, message, executedAt, duration)
}

// And formatWarning.
func formatWarning(hostGlobalShort, server, message, eventShort, executedAt, duration string) string {

	warningMarker := coloredMarker(ColorYellow)
	return fmt.Sprintf(
		"[%s][%s] %s %s %s Warning %s\nExecuted At: %s | Duration: %s\n",
		hostGlobalShort, eventShort, warningMarker, server, warningMarker, message, executedAt, duration)
}

type TaskResult struct {
	Host       string `json:"host"`
	Task       string `json:"task"`
	Module     string `json:"module"`
	Success    bool   `json:"success"`
	Result     string `json:"result,omitempty"`
	Error      string `json:"error,omitempty"`
	EventID    string `json:"event_id"`
	GlobalID   string `json:"global_id"`   // Unique per host session
	ExecutedAt string `json:"executed_at"` // Timestamp when task finished
	Duration   string `json:"duration"`    // Duration of the task execution
}

type HostSummary struct {
	Total       int
	Failed      int
	Unreachable bool
}

func isConnectionError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "connection failed")
}

// safeExecute wraps module.Execute with a deferred recover to catch panics.
// Note: The parameter type has been changed to map[string]string.
func safeExecute(ctx context.Context, module core.Module, host inventory.Host, params map[string]string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during module execution: %v", r)
		}
	}()
	result, err = module.Execute(ctx, host, params)
	return result, err
}

// executeWithRetry attempts a task until retries are exhausted or success, using safeExecute.
func executeWithRetry(ctx context.Context, module core.Module, host inventory.Host, task tasks.Task) (string, error) {
	maxRet := taskRetries(task)
	delay := taskRetryDelay(task)
	var result string
	var err error
	var attemptsUsed int
	for attempt := 1; attempt <= maxRet; attempt++ {
		result, err = safeExecute(ctx, module, host, task.Params)
		attemptsUsed = attempt
		if err == nil {
			return result, nil
		}
		if isConnectionError(err) {
			return "", fmt.Errorf("after %d SSH attempts, %w", attemptsUsed, err)
		}
		logrus.WithFields(logrus.Fields{
			"host":    host.Name,
			"task":    task.Name,
			"module":  task.Module,
			"attempt": attempt,
		}).Warnf("Task execution failed: %v", err)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
		}
	}
	return result, err
}

func taskRetries(task tasks.Task) int {
	if task.MaxRetries > 0 {
		return task.MaxRetries
	}
	return defaultMaxRetries
}

func taskRetryDelay(task tasks.Task) time.Duration {
	if task.RetryDelay != "" {
		d, err := time.ParseDuration(task.RetryDelay)
		if err == nil {
			return d
		}
		logrus.Warnf("Invalid retry_delay for task %s: %v, using default", task.Name, err)
	}
	return defaultRetryDelay
}

func loadInventory(path string) (inventory.Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return inventory.Inventory{}, err
	}
	var inv inventory.Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return inv, err
	}
	return inv, nil
}

func loadTasks(path string) ([]tasks.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t []tasks.Task
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return t, nil
}

func main() {
	// Global recovery: catch any panic in main.
	defer func() {
		if r := recover(); r != nil {
			logrus.Fatalf("Unhandled panic in main: %v", r)
		}
	}()

	invPath := flag.String("inventory", "configs/inventory.yml", "Path to inventory file")
	tasksPath := flag.String("tasks", "configs/tasks.yml", "Path to tasks file")
	outputJSON := flag.Bool("output-json", false, "Output results in JSON format")
	parallel := flag.Bool("parallel", true, "Launch tasks on servers concurrently. Use false for sequential execution.")
	flag.Parse()

	interactive := !(*outputJSON)

	if *outputJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	} else {
		logrus.SetOutput(io.Discard)
	}
	logrus.SetLevel(logrus.InfoLevel)

	inv, err := loadInventory(*invPath)
	if err != nil {
		logrus.Fatalf("Failed to load inventory: %v", err)
	}
	taskList, err := loadTasks(*tasksPath)
	if err != nil {
		logrus.Fatalf("Failed to load tasks: %v", err)
	}

	moduleRegistry := modules.RegisterModules()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	var resMutex sync.Mutex
	var results []TaskResult

	printJSONResult := func(res TaskResult) {
		j, err := json.Marshal(res)
		if err == nil {
			fmt.Println(string(j))
		}
	}

	// processHost now recovers from any panic during host processing.
	processHost := func(host inventory.Host) {
		defer func() {
			if r := recover(); r != nil {
				errMsg := fmt.Sprintf("Recovered from panic in host %s: %v", host.Name, r)
				printColored(ColorRed, errMsg)
				// Optionally, record a TaskResult for the host panic.
			}
		}()

		hostName := host.Name
		// Generate a unique global id per host session.
		hostRunningID := uuid.New().String()
		hostGlobalShort := hostRunningID[:8]
		abortHost := false
		for _, task := range taskList {
			if abortHost {
				break
			}
			// Generate an event id for the current task.
			eventID := uuid.New().String()
			eventShort := eventID[:8]

			// Uncomment the next line to print launching message.
			// fmt.Print(formatLaunchingTask(hostGlobalShort, hostName, task.Name, eventShort))

			module, exists := moduleRegistry[task.Module]
			if !exists {
				msg := fmt.Sprintf("Module %s not found for task %s", task.Module, task.Name)
				executedAt := time.Now().Format(time.RFC3339)
				duration := "0s"
				tr := TaskResult{
					Host:       hostName,
					Task:       task.Name,
					Module:     task.Module,
					Success:    false,
					Error:      msg,
					EventID:    eventID,
					GlobalID:   hostRunningID,
					ExecutedAt: executedAt,
					Duration:   duration,
				}
				resMutex.Lock()
				results = append(results, tr)
				resMutex.Unlock()
				if *outputJSON {
					printJSONResult(tr)
				} else if interactive {
					fmt.Print(formatWarning(hostGlobalShort, hostName, msg, eventShort, executedAt, duration))
				}
				continue
			}

			// Record the start time.
			startTime := time.Now()
			result, err := executeWithRetry(ctx, module, host, task)
			duration := time.Since(startTime).String()
			executedAt := time.Now().Format(time.RFC3339)
			var tr TaskResult
			if err != nil {
				msg := fmt.Sprintf("Task %s failed: %v", task.Name, err)
				tr = TaskResult{
					Host:       hostName,
					Task:       task.Name,
					Module:     task.Module,
					Success:    false,
					Error:      msg,
					EventID:    eventID,
					GlobalID:   hostRunningID,
					ExecutedAt: executedAt,
					Duration:   duration,
				}
				resMutex.Lock()
				results = append(results, tr)
				resMutex.Unlock()
				if *outputJSON {
					printJSONResult(tr)
				} else if interactive {
					fmt.Print(formatError(hostGlobalShort, hostName, msg, eventShort, executedAt, duration))
				}
				if isConnectionError(err) {
					abortHost = true
				}
				continue
			}
			tr = TaskResult{
				Host:       hostName,
				Task:       task.Name,
				Module:     task.Module,
				Success:    true,
				Result:     result,
				EventID:    eventID,
				GlobalID:   hostRunningID,
				ExecutedAt: executedAt,
				Duration:   duration,
			}
			resMutex.Lock()
			results = append(results, tr)
			resMutex.Unlock()
			if *outputJSON {
				printJSONResult(tr)
			} else if interactive {
				fmt.Print(formatSuccess(hostGlobalShort, hostName, task.Name, result, eventShort, executedAt, duration))
			}
		}
	}

	if *parallel {
		for _, host := range inv.Hosts {
			wg.Add(1)
			go func(h inventory.Host) {
				defer wg.Done()
				processHost(h)
			}(host)
		}
	} else {
		for _, host := range inv.Hosts {
			processHost(host)
		}
	}
	wg.Wait()

	if *outputJSON {
		finalSummary, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(finalSummary))
		return
	}

	summary := make(map[string]*HostSummary)
	for _, res := range results {
		if _, ok := summary[res.Host]; !ok {
			summary[res.Host] = &HostSummary{}
		}
		summary[res.Host].Total++
		if !res.Success {
			summary[res.Host].Failed++
			if isConnectionError(errors.New(res.Error)) {
				summary[res.Host].Unreachable = true
			}
		}
	}
	fmt.Println()
	printColored(ColorBlue, "Summary per server:")
	for host, sum := range summary {
		status, color := "OK", ColorGreen
		if sum.Unreachable {
			status, color = "UNREACHABLE", ColorRed
		} else if sum.Failed > 0 {
			status, color = "FAILED", ColorRed
		}
		printColored(color, fmt.Sprintf("Server: %s - Total: %d, Success: %d, Failed: %d, Status: %s",
			host, sum.Total, sum.Total-sum.Failed, sum.Failed, status))
	}
	printColored(ColorBlue, "All tasks completed.")
}
