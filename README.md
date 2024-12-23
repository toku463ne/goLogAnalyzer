# goLogAnalyzer
Group similar logs, display trends for each log group, and help identify abnormal behaviors in the logs.  
  
## Base Concept  
This application counts the occurrences of all terms in log files.  
Terms that appear less than R times will be replaced with an asterisk ("*").  
After the replacement, identical log lines will be grouped together, forming a "log group".  
  
## Basical usage  
### Step1
Prepare a YAML file with minimal contents as shown below:  
It is important to pick up <timestamp> and <message> from the log.  
```yaml
dataDir: "myLogDataDir"
logPath: "/var/log/syslog*"
logFormat: '^(?P<timestamp>\w{3} \d{1,2} \d{2}:\d{2}:\d{2}) (?P<message>.*)$'
keepPeriod: 60 # terms to keep the analyzed data
unitSecs: 86400 # daily
timestampLayout: "Jan 2 15:04:05"
```  
  
Save it as `myConfig.yaml`.  
  
### Step2
Test the YAML file with a log line using the following command:
```sh
logan test -c myConfig.yaml -line "Dec 22 00:00:05 br-004-091 systemd[1]: logrotate.service: Succeeded"
```
If the YAML file is correct, you will see the following output:
```
the line parsed as:
timestamp:  2024-12-22
message:  br-004-091 systemd[1]: logrotate.service: Succeeded
```  
  
### Step3
Feed logs  
```sh
logan feed -c myConfig.yaml 
```
