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
logan [run|clean|stats|topN|report] OPTIONS

logan -help:
        Shows this help

logan run:
	Calculate rarity score of each log records and show the "rare" records.
	Run "logan run -help" for details.

logan report:
	Reads params from json config file.
	Run "logan report -help" for details.

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

```
Usage of run: 
      -blockPerFile
            If create blocks per files
      -blockSize int
            Max number of lines in a block
      -d string
            Directory to save the analyzation data
      -dateLayout string
            Layout of datetime in the log
      -dateStart int
            Start position of datetime in the log starting from 0
      -f string
            Log file(regex) to analyze. Supports data from pipe
      -g float
            Gap rate from average
                              Log records with rarity score whose gap if higher that this value will be showed.
      -maxBlock int
            max blocks to save logs
      -maxItemBlock int
            max blocks to save terms
      -n int
            Top N rare records to show (default 10)
      -nRareTerms int
            Top N rare terms to display (default 20)
      -save
            Update the data without asking
      -scoreNSize int
            How to calculate the score.
            1:simple average
            2:average of top scoreNSize terms in a record (default 10)
      -scoreStyle int
            How to calculate the score.
            1:simple average
            2:average of top scoreNSize terms in a record (default 3)
```
  
```
Usage of clean:
  -d string
        Directory to save the analyzation data
```
  
```
Usage of stats:
  -d string
        Directory to save the analyzation data
```

```
Usage of topN:
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

  -v bool  
        Show score of items in the log record
```

```
Usage of report:
  -c string
        Path of the config file (JSON)
```  
  
## Examples
Run in most simple way.  
This exampe do not save any data.
```
logan run -f /var/log/syslog
tail -n 1000 /var/log/syslog | logan run # supports pipe
```
  
Run collecting data.  
Recommended when analyzing huge size log files, or if next time you want to restart from the point you finished.   
```
logan run -f '/var/log/syslog*' -d data  # supports regex
```
and next time you can run the way below
```
logan run -d data
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
    
Create HTML reports  
Prepare a json file. (refer to configsample.json)  
```
logan report -c configsample.json
```  
