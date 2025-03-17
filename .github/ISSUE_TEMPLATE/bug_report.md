---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**describe the bug**  
when a host is unreachable, opereta continues to retry tasks instead of aborting further attempts on that host. this leads to multiple error messages and delays, even though the expected behavior is to stop processing tasks for unreachable hosts.

**to reproduce**  
1. add a host in your inventory file with an incorrect address or port (making it unreachable).  
2. create a task that executes a simple command (for example, "echo hello") on that host.  
3. run opereta with your inventory and tasks file (e.g. `./opereta -inventory configs/inventory.test -parallel=false`).  
4. observe that opereta retries the task multiple times and logs repeated error messages (e.g. "after 1 task attempts, SSH connection failed...").

**expected behavior**  
when a host is unreachable, opereta should mark that host as unreachable and skip further tasks for that host without further retries.

**screenshots**  
if applicable, attach a screenshot of the terminal output showing the repeated error messages and retries.

**desktop (please complete the following information):**  
- os: linux (or specify your os)  
- terminal: (e.g. gnome-terminal, iterm, etc.)  
- version: (provide version if applicable)

**smartphone (please complete the following information):**  
- device: not applicable  
- os: not applicable  
- browser: not applicable  
- version: not applicable

**additional context**  
this issue appears to occur when a connection error is detected, but opereta does not abort further tasks on that host. this may lead to unwanted delays and excessive logging. any help or suggestions for a proper fail-fast approach on unreachable hosts would be appreciated.
