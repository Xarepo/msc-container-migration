#!/bin/sh
DUMP_DIR=/dumps
DUMPS=$(ls -l $DUMP_DIR | tail -n +2 | awk '{print $9}')
FULL_DUMPS=$(echo "$DUMPS" | sed -n '/d[0-9]\+/p')
PRE_DUMPS=$(echo "$DUMPS" | sed -n '/p[0-9]\+/p')

# Sums the size of the dumps. Uses bytes stored rather than entire blocks.
function sum_dumps() {
	local SUM=0
	for dump in $1 ; do
		files=$(ls -l $DUMP_DIR/$dump | tail -n +2 | awk '{print $9}')
		for file in $files ; do
			size=$(stat $DUMP_DIR/$dump/$file | sed -n '2p' | awk '{print $2}')
			SUM=$((SUM + size))
		done
	done
	echo $SUM
}

SUM_PRE=$(sum_dumps "$PRE_DUMPS")
NUM_PRE=$(echo "$PRE_DUMPS" | wc -l)
AVG_PRE=$((SUM_PRE/NUM_PRE))

SUM_FULL=$(sum_dumps "$FULL_DUMPS")
NUM_FULL=$(echo "$FULL_DUMPS" | wc -l)
AVG_FULL=$((SUM_FULL/NUM_FULL))

echo SUM_PRE=$SUM_PRE
echo NUM_PRE=$NUM_PRE
echo AVG_PRE=$AVG_PRE

echo

echo SUM_FULL=$SUM_FULL
echo NUM_FULL=$NUM_FULL
echo AVG_FULL=$AVG_FULL
