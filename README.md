# goLogAnalyzer
A simple tool to analize log files.  
It supports plain text and gzip formatted log files.
  
## Overview
Find log records which are rare.


Once you run this tool, the result will be saved and next time the tool will start analyzation from where it finished the last time.  
  
## Installation
```
./install.sh  # linux  
./install.bat  # windows  
```
  
## How to use
- **loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]**
  Starts log analyzation.  
  ```
	-f LOGPATH:   
		Path of the logfile (can use regex)  
	-v verbose  
	-g GAPVALUE:  
		Only shows logs which have gap from average score.  
		Default is 0.8  
		0 is the average.  
		1 is 1 deviation width from the average.  
		The score is calculated as below and indicates how rare the log record is.  
		term score: log10((count of all terms)/(count of the term)) + 1  
		log record score: average of term scores in the log record  
		* Count is calculated at the point the log record appeared.  
	-d DATADIR:  
		Directory to save the analyzation data.  
		For large log files, using this option is recomended.
		Otherwise goLogAnalyzer may use much memory.
		This data will be also used in the next time execution  
	-s SEARCH_KEYS:  
		key word to search (can use regex)  
	-x EXCLUDE_KEYS:  
		key word to exclude (can use regex)  
    ```  

- **loganal clean -d DATADIR**
    ```
  Cleans up the analyzation data in previous analysis  
    ```  

- **loganal frq -f LOGPATH [-m MIN_SUPPORT] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]**  
  Shows Closed Frequent Itemset order by the support  
    ```
	-f LOGPATH: Path of the logfile  
	-m MIN_SUPPORT: minimum support of closed frequent item sets  
	-s SEARCH_KEYS: key word to search (can use regex)  
	-x EXCLUDE_KEYS: key word to exclude (can use regex)  
    ```