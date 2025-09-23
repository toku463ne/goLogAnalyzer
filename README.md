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
This is provided as CSV files. (metrics format)
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
`<patternKey>` is a key to identify a session.  
`<from>` and `<to>` are user defined keys to group the patterns.
Then run the command
```
logan patterns -c myConfig.yaml 
```
This will show up the patterns found in the log file.  
Output example:
```
175716601400021 175716601400022 175716601400023 175716601400023 175716601400024 175716601400025 175716601400026 175716601400027 175716601400028 175716601400029 175716601400030 175716601400031 175716601400032 175716601400033 175716601400034 175716601400035 175716601400026 175716601400036 175716601400028 175716601400029 175716601400037 175716601400038 175716601400039 175716601400034 175716617300004 175716601400040 175716601400041 175716601400042 175716601400043 175716601400044 175716601400045 175716601400046 175716601400042 175716601400047 175716601400043 175716601500048 175716601500049 175716601500050 175716617300001 175716617300002 175716617300003 175716617300004 175716617300005 175716617300004 175716617300006 175716617300007 175716617300008 175716617300009 175716617300010 175716617300011 175716617300012 
=> total 1
to:0111117017, from:05000002360: {startEpoch: 2025/09/06 22:40:14, count: 1}
---
TBLV1 CALL: CTBCMCLeg::*( LegId=* Type CALL, NAP NAPS_BR_PRO_FE, calling/* */* )
TBLV1 CALL: Unknown Entity * [*: *-*] CTBCAFBridge: CTBCAFCallLeg::CTBCAFCallLeg
TBLV1 CALL: Leg * [*: *-*] *: *
TBLV1 CALL: Leg * [*: *-*] *: *
TBLV3 CALL: [*: *-*] CTBCAFBridge: *: Leg *, MappingId * (calling 05000002360, called 0111117017, NAP NAPS_BR_PRO_FE):
TBLV3 CALL: [*: *-*] CTBCAFBridge: Local SDP:
TBLV3 CALL: [*: *-*] CTBCAFBridge: v=0
TBLV3 CALL: [*: *-*] CTBCAFBridge: o=- 0 0 IN IP4 1.2.3.111
TBLV3 CALL: [*: *-*] CTBCAFBridge: s=-
TBLV3 CALL: [*: *-*] CTBCAFBridge: t=0 0
TBLV3 CALL: [*: *-*] CTBCAFBridge: m=audio * RTP/AVP 0 8 18 101
TBLV3 CALL: [*: *-*] CTBCAFBridge: c=IN IP4 1.2.3.111
TBLV3 CALL: [*: *-*] CTBCAFBridge: a=rtpmap:101 telephone-event/8000
TBLV3 CALL: [*: *-*] CTBCAFBridge: a=fmtp:101 0-15,32-36
* CALL: [*: *-*] CTBCAFBridge:
TBLV3 CALL: [*: *-*] CTBCAFBridge: Peer SDP:
TBLV3 CALL: [*: *-*] CTBCAFBridge: v=0
TBLV3 CALL: [*: *-*] CTBCAFBridge: o=- 1754469684 1754469685 IN IP4 1.2.3.209
TBLV3 CALL: [*: *-*] CTBCAFBridge: s=-
TBLV3 CALL: [*: *-*] CTBCAFBridge: t=0 0
TBLV3 CALL: [*: *-*] CTBCAFBridge: m=audio * RTP/AVP 0 8 * 18
TBLV3 CALL: [*: *-*] CTBCAFBridge: c=IN IP4 1.2.3.209
TBLV3 CALL: [*: *-*] CTBCAFBridge: a=fmtp:18 *=no
* CALL: [*: *-*] CTBCAFBridge:
TBLV1 CALL: Leg * [*: *-*] CTBCAFBridge: *::*
TBLV1 CALL: [*: *-*] CTBCAFCallBehaviorBusyTone: OnInitCallDone
TBLV1 CALL: [*: *-*] CTBCAFBridge: CTBCAFCallFlow::OnInitCallDone
TBLV1 CALL: Leg * [*: *-*] CTBCAFCallBehaviorBridgeCdr: OnLegEvent type = 4
TBLV1 CALL: Leg * [*: *-*] CTBCAFCallBehaviorRouting: Route: Now routing for a new outgoing call...
TBLV1 CALL: Leg * [*: *-*] CTBCAFCallBehaviorRouting: ProcessRouteResult: ReasonCode 219
TBLV1 CALL: Leg * [*: *-*] *: * check...
```
  
## more details
Run
```
logan groups -h
```