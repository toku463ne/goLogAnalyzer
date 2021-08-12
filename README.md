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
logan [rar|clean|stats|topN] OPTIONS

logan -help:
        Shows this help

logan rar:
        Calculate rarity score of each log records and show the "rare" records.
        Run "logan rar -help" for details.

logan clean:
        Cleanups all statistics data.
        Run "logan clean -help" for details.

logan stats:
        Shows the statistics according the saved data.
        Run "logan stats -help" for details.

logan topN:
        Shows the top N rare records
        Run "logan topN -help" for details.

  ```
  
Usage of rar:
  ```
Usage of rar:
  -d string
        Directory to save the analyzation data

  -f string
        Log file(regex) to analyze. Supports data from pipe

  -g float
        Gap rate from average
        Log records with rarity score whose gap if higher that this value will be showed. 
        (default 0.5)

  -linesInBlock int
        lines in block (default 10000)

  -maxBlock int
        max blocks (default 100)

  -maxItemBlock int
        max blocks for items (default 1000)

  -n int
        max lines to process

  -s string
        key word to search

  -x string
        key word to exclude

  -save
        Update the data without asking
  
  ```
  
Usage of clean:
  ```
  -d string
        Directory to save the analyzation data
  ```
  
Usage of stats:
  ```
  -d string
        Directory to save the analyzation data

  -n int
        Number of history to show (default 5)
  ```

Usage of topN:
  ```
  -d string
        Directory to save the analyzation data

  -n int
        Top N rare records to show (default 10)

  -s string
        key word to search

  -start string
        Start date to collect stats %Y-%m-%d format

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
  
Check the result from saved data  
```
logan stats -d data
```  

To exclude phrases in the result  
```
logan stats -d data -x "System clock|The system clock|IKE"
```  
  
To show top N rare records   
```
logan topN -d data
```  
    
