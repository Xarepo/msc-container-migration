#!/bin/sh

SOURCE_DIR=$1

# Clear result files first
rm -f pre.txt full.txt

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

	AVG_PRE=$(cat $RESULT_FILE | sed -n '/AVG_PRE/p' | awk '{print $1}' | cut -d'=' -f2- | sed 's/..$//')
	AVG_FULL=$(cat $RESULT_FILE | sed -n '/AVG_FULL/p' | awk '{print $1}' | cut -d'=' -f2- | sed 's/..$//')

	echo $DB_SIZE $AVG_PRE >> pre.txt
	echo $DB_SIZE $AVG_FULL >> full.txt
done
