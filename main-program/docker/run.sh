#!/bin/bash
if [ -z $MASTER ]
then
	echo "starting with default configuration"
    /go/bin/mmq -f configuration.json    
else
	echo "loading from master"
	rm configuration.json
	/go/bin/mmq -l $MASTER
fi