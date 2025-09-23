# goLogAnalyzer
Group similar logs, display trends for each log group, and help identify abnormal behaviors in the logs.  
  
## Base Concept  
This application counts the occurrences of all terms in log files.  
Terms that appear less than R times will be replaced with an asterisk ("*").  
After the replacement, identical log lines will be grouped together, forming a "log group".  
  
## Basical usage  
For simple analyzation to not so large logs
```
logan -f /var/log/syslog
```
This will show up the top 20 frequent log groups.  
  
You can filter words by "-s" option
```
logan -f /var/log/syslog -s error
```
  
You can exclude words by "-x" option
```
logan -f /var/log/syslog -x systemd
```
  
If you want to do more complex tasks like separating timestamps, grouping by sessionIds etc., then analyze with conf file 

## Analyze with conf file
### Step1
Prepare a YAML file with minimal contents as shown below:  
It is important to pick up <timestamp> and <message> from the log.  
```yaml
dataDir: "/tmp/myLogDataDir"
logPath: "/var/log/syslog*"
logFormat: '^(?P<timestamp>\w{3} \d{1,2} \d{2}:\d{2}:\d{2}) (?P<message>.*)$'
keepPeriod: 60 # terms to keep the analyzed data
unitSecs: 86400 # daily
timestampLayout: "Jan 2 15:04:05"
```
(*) Note that meta data will be saved at `dataDir`.    
  
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
# example  
Group the log lines with default grouping rate
```sh
logan -f 'testdata/loganal/sample50*' -N 5
Log Groups
==========
Group ID   Count      Text
1730671576000000005 10         *:00:00] Com1, grpe10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpc20 (uniq)*
1730671576000000001 10         *:00:00] Com1, grpa10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpa20 (uniq)*
1730671576000000002 10         *:00:00] Com1, grpb10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpa20 (uniq)*
1730671576000000003 10         *:00:00] Com1, grpc10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpb20 (uniq)*
1730671576000000004 10         *:00:00] Com1, grpd10 Com2 (uniq)* grpa50 (uniq)* <coM3> (uniq)* grpb20 (uniq)*
```
  
### Step4 (optional)
Remove the meta data
```
logan clean -c myConfig.yaml
```
  
## commands
### feed
`feed` command analyzes logs and create meta data.  
For large log files, the analyzation task may take time.
Once you create meta data, other commands like `history`, `groups`, `patterns` will be fast.
```
logan feed -c myConfig.yaml
```
  
### history
Devide log groups per `unitSecs` and saves in timestamp order.  
You can the log group count per `unitSecs`.  
For access logs, this would help to see what kind of accesses are frequent in a timeline.  
For example, you can list the top 5 log groups by 
```
logan -c myConfig.yaml
```
and then check the history by the groupId
```
logan history -c myConfig.yaml -groupId 123456789 
```
or output to CSV files
```
logan history -c myConfig.yaml -o /tmp/logancsv
```
  
### pattern
Prepare a config file with `patternDetectionMode` and `patternKeyRegexes`.
```
dataDir: "{{ HOME }}/logantests/Test_sbc_gateway/data"
logPath: "../../testdata/loganal/sbc_gateway.log"
logFormat: '^(?P<timestamp>\d{1,2}th, \d{2}:\d{2}:\d{2}\.\d{3}\+\d{4}) (?P<message>.*)$'
timestampLayout: "02th, 15:04:05.000-0700"
minMatchRate: 0
termCountBorder: 100
countBorder: 100
blockSize: 1000
unitSecs: 3600
separators: ' "''\\,;[]<>{}=()|:&?/+!@-'
patternDetectionMode: "relations"
patternKeyRegexes:
  - 'TBLV1 CALL: CTBCMCLeg::Construct.* LegId=(?P<patternKey>\w+) Type CALL, NAP .* calling/called (?P<from>\w+)/(?P<to>\w+) .*'
ignoreRegexes:
  - '^0x.*'
```