#!/bin/sh
DUMP_DIR=/dumps
DUMPS=$(ls -l $DUMP_DIR | tail -n +2 | awk '{print $9}')
FULL_DUMPS=$(echo "$DUMPS" | sed -n '/d[0-9]\+/p')
PRE_DUMPS=$(echo "$DUMPS" | sed -n '/p[0-9]\+/p')

# Sums the size of the dumps. Uses bytes stored rather than entire blocks.
sum_dumps() {
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
NUM_PRE=$([ -n "$PRE_DUMPS" ] && echo "$PRE_DUMPS" | wc -l || echo 0)
AVG_PRE=$([ $NUM_PRE -ne 0 ] && echo $((SUM_PRE/NUM_PRE)) || echo 0)

SUM_FULL=$(sum_dumps "$FULL_DUMPS")
NUM_FULL=$([ -n "$FULL_DUMPS" ] && echo "$FULL_DUMPS" | wc -l || echo 0)
AVG_FULL=$([ $NUM_FULL -ne 0 ] && echo $((SUM_FULL/NUM_FULL)) || echo 0)


present_bytes(){
	bytes=$1
	echo "$bytes"B, $(numfmt --to=iec-i --suffix=B $bytes)
}

echo SUM_PRE=$(present_bytes $SUM_PRE)
echo AVG_PRE=$(present_bytes $AVG_PRE)
echo NUM_PRE=$NUM_PRE

echo

echo SUM_FULL=$(present_bytes $SUM_FULL)
echo AVG_FULL=$(present_bytes $AVG_FULL)
echo NUM_FULL=$NUM_FULL
