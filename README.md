# goLogAnalyzer
A simple tool to analize log files.  
It supports plain text and gzip formatted log files.  
Tested on linux, but also works in Windows.
  
## Overview
Find log records which are rare.
  

If you specify a datadir, the result will be saved, and next time the tool will start analyzation from where you finished.  
  
## Installation
```
./install.sh  
```  

  
## How to use
- **logan rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]**  
  Starts log analyzation.  
  ```
	-f LOGPATH:   
		Path of the logfile (can use regex)  
	-v verbose: 
		Show debug logs  
	-g GAPVALUE:(default 0.8)  
		Defines the threshold of "rarity".
		Log records with "rarity" score higher than this value will be showed up.
		0 is the average.  
		1 is 1 deviation width from the average.  
	-d DATADIR:   
		Directory to save the analyzation data.  
		For large log files, using this option is recomended.
		Otherwise gologanyzer may use much memory.
		This data will be also used in the next time execution  
	-s SEARCH_KEYS:  
		key word to search (can use regex)  
	-x EXCLUDE_KEYS:  
		key word to exclude (can use regex)  
    ```  

- **logan clean -d DATADIR**
    ```
  Cleans up the analyzation data in previous analysis  
    ```  

- **logan frq -f LOGPATH [-m MIN_SUPPORT] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]**  
  Shows Closed Frequent Itemsets order by the support  
    ```
	-f LOGPATH: Path of the logfile  
	-m MIN_SUPPORT: minimum support of closed frequent item sets  
	-s SEARCH_KEYS: key word to search (can use regex)  
	-x EXCLUDE_KEYS: key word to exclude (can use regex)  
    ```
  
- **logan test [-f LOGPATH] [-d DATADIR] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]**  
  Same as "logan rar" but will shows all log records
  
  

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
  
  
If you think there is too much output or too less output, then change the GAPVALUE.  
Reducing the GAPVALUE means that log records with less "rarity" will show up.  
So you can get more outputs by reducing the GAPVALUE. (default 0.8)
```
logan rar -f /var/log/syslog -g 0.5
```  
  
If you want to see all log records with score  
```
logan test -f /var/log/syslog 
```
  
