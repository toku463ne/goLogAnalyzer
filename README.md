# goLogAnalyzer
A simple tool to analize log files.  
It supports plain text and gzip formatted log files.  
Tested on linux and Windows.
  
## Overview
Find log records which are rare.
  

If you specify a datadir, the result will be saved, and next time the tool will start analyzation from where you finished.  
  
## Installation
```
./install.sh  
```  

  
## How to use
  ```
logan [rar|clean|stats|test|frq] OPTIONS

logan -help:
        Shows this help

logan rar:
        Calculate rarity score of each log records and show the "rare" records.
        Run "logan rar -help" for details.

logan clean:
        Cleanups all statistics data.
        Run "logan clean -help" for details.

logan stats:
        Shows the statistics according the data in the last execution.
        Run "logan stats -help" for details.

logan test:
        Shows all log records with the score gap.
        Run "logan test -help" for details.

logan frq:
        Shows the closed frequent itemsets order by the supports.
        Only calculate at most 10000 records.
  ```
  
Usage of rar:
  ```
  -a    show old results too
  -d string
        Directory to save the analyzation data
  -f string
        Log file(regex) to analyze. Supports data from pipe
  -g float
        Gap rate from average
                        Log records with rarity score whose gap if higher that this value will be showed.
  -linesInBlock int
        lines in block (default -1)
  -maxBlock int
        max blocks (default -1)
  -n int
        max lines to process
  -r float
        Log records with top "rarity score" will be showed.
                        Default is 0.00005 (5 rare record out of 100000 records will be showed)
  -s string
        key word to search
  -save
        Update the data without asking
  -v    show debug logs
  -x string
        key word to exclude
  ```
  
Usage of clean:
  ```
  -d string
        Directory to save the analyzation data
  -v    show debug logs
  ```
  
Usage of stats:
  ```
  -d string
        Directory to save the analyzation data
  ```
  
Usage of frq:
  ```
  -f string
        Text file to analyze
  -m int
        min support
  -s string
        key word to search
  -v    show debug logs
  -x string
        key word to exclude
  ```

## Examples
Run in most simple way.  
This exampe do not save any data.
```
logan rar -f /var/log/syslog
tail -n 1000 /var/log/syslog | logan rar # supports pipe
```
  
Run collecting data.  
Recommended when analyzing huge size log files, or if next time you want to restart from the point you finished.   
```
logan rar -f '/var/log/syslog*' -d data  # supports regex
```
and next time you can run the way below
```
logan rar -d data
```
  
  
If you think there is too much output or too less output, then change the RARITY_RATE.  
Reducing the RARITY_RATE means that log records with more "rarity" will show up.  
So you can get more outputs by increasing the RARITY_RATE. (default 0.0001)
```
logan rar -f /var/log/syslog -r 0.0002
```  
  
If you want to see all log records with score  
```
logan test -f /var/log/syslog 
```
  
