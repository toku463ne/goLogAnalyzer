# goLogAnalyzer
A simple tool to analize log files.  
It supports plain text and gzip formatted log files.
  
## Overview
Find log records which are
- **rare**  
    "Rarity" of a log record is defined by sum of IDF scores of the terms in the log record.   
- **absent after appeared frequently**  
    "frequency" is applied to "closed frequent item sets".  
    This tool devide the log files in some blocks and calculate "closed frequent item sets" per blocks.  
    If a "closed frequent item set" does not appear in the new block, the "closed frequent item set" is considered "absent".  
    If the "absence" of the "closed frequent item set" is "rare", it is considered "absent after appeared frequently".  
  
Once you run this tool, the result will be saved and next time the tool will start analyzation from where it finished the last time.  
  
## Installation
```
./install.sh
```
  
## How to use
Copy config.ini.sample and edit.
```
[LogFile]
linesInBlock = 1000
maxBlocks = 10
rootDir = ~/loganal ## files will be saved under this directory
logName = syslog ## label of the logfile
logPathRegex = /var/log/syslog* 
rarityThreshold = 0.8  
frequencyCheck = true
frequencyThreshold = 0.5
minSupportPerBlock = 0.02
absencyThreshold = 0.5
```

and run the command below
```
loganal -c config.ini
```
or if the file name of the config is "config.ini", just run 
```
loganal
```
