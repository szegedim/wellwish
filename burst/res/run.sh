#!/bin/bash

while true;
do
    export BK=`curl -X GET http://127.0.0.1:7777/idle?apikey=WABIWLVZLWWGTGXQSOFSPHWBGNTQRPQUJYYPEGJITNSCGQPTOZHJRDWIMXHDQCYLMTFOFNFUHGTPPJGCBGOBNTVWAPER`
    docker run --rm -d --name $BK -e BURSTKEY=$BK wellwish
    sleep 30
    docker kill $BK
done
