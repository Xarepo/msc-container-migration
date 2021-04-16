#!/bin/sh

SOURCE_DIR=$1

rm -f data.csv

echo dbsize,pre,full >> data.csv
for dir in $(ls -A1 $SOURCE_DIR) ; do
	RESULT_FILE=$SOURCE_DIR/$dir/result.log
	# Get the size of the redis database 
	DB_SIZE=$(grep -o 'POST_RUN_HOOK=".*"' $RESULT_FILE | \
		cut -d'=' -f2- | awk '{print $NF}' | tr -d '"')
	if [ -z $DB_SIZE ]; then
		# If the string is empty then there was no post run hook, i.e. the database
		# is empty.
		DB_SIZE=0
	fi

	# Display results as kilobytes
	AVG_PRE=$(cat $RESULT_FILE | \
		sed -n '/AVG_PRE/p' | \
		awk '{print $1}' | \
		cut -d'=' -f2- | \
		sed 's/..$//' | \
		xargs -I _ echo "_ / (2^10)" | \
		bc)
	AVG_FULL=$(cat $RESULT_FILE | \
		sed -n '/AVG_FULL/p' | \
		awk '{print $1}' | \
		cut -d'=' -f2- | \
		sed 's/..$//' | \
		xargs -I _ echo "_ / (2^10)" | \
		bc)

	echo "$DB_SIZE,$AVG_PRE,$AVG_FULL" >> data.csv
done
