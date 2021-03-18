#!/bin/sh

SOURCE_DIR=$1

for dir in $(ls -A1 $SOURCE_DIR) ; do
	# Get the size of the redis database 
	DB_SIZE=$(grep -o 'POST_RUN_HOOK=".*"' $SOURCE_DIR/$dir/results | \
		cut -d'=' -f2- | awk '{print $NF}' | tr -d '"')
	if [ -z $DB_SIZE ]; then
		# If the string is empty then there was no post run hook, i.e. the database
		# is empty.
		DB_SIZE=0
	fi

	DOWNTIME=$(cat $SOURCE_DIR/$dir/results | sed -n '/AVG_DOWNTIME/p' | awk '{print $2}')
	TOT_TIME=$(cat $SOURCE_DIR/$dir/results | sed -n '/AVG_TOT_MIG_TIME/p' | awk '{print $2}')

	echo $DB_SIZE $DOWNTIME >> downtime.dat
	echo $DB_SIZE $TOT_TIME >> tot_time.dat
done
