#!/bin/sh
DUMP_DIR=/dumps
DUMPS=$(ls -l $DUMP_DIR | tail -n +2 | awk '{print $9}')

SUM=0
for dump in $DUMPS ; do
	files=$(ls -l $DUMP_DIR/$dump | tail -n +2 | awk '{print $9}')
	for file in $files ; do
		size=$(stat $DUMP_DIR/$dump/$file | sed -n '2p' | awk '{print $2}')
		SUM=$((SUM + size))
	done
done

echo SUM: $SUM
