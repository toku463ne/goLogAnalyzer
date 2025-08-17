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
Analyze your log  
# example 1  
Group the log lines with default grouping rate
```sh
logan groups -f 'testdata/loganal/sample50*' -N 5
Log Groups
==========
Group ID   Count      Text
1730671576000000005 10         *:00:00] Com1, grpe10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpc20 (uniq)*
1730671576000000001 10         *:00:00] Com1, grpa10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpa20 (uniq)*
1730671576000000002 10         *:00:00] Com1, grpb10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpa20 (uniq)*
1730671576000000003 10         *:00:00] Com1, grpc10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpb20 (uniq)*
1730671576000000004 10         *:00:00] Com1, grpd10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpb20 (uniq)*
```
