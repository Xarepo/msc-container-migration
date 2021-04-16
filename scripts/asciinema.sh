OUTPUTNAME=$1 
rm -f $OUTPUTNAME.cast
tmux new-session -d -s msc-asciinema
asciinema rec --command "tmux attach -t msc-asciinema \; \
	new-window '\
		docker run \
			--name msc-asciinema \
			--env-file ./.env \
			--rm --privileged \
			-v $(pwd)/rootfs:/app/rootfs \
			-v $(pwd)/config.json:/app/config.json \
			msc run msc && \
		sleep 10' \; \
	split-window -h 'sleep 2 && \
		docker run \
			--name msc2-asciinema \
			--env-file ./.env \
			--rm --privileged \
			-v $(pwd)/rootfs:/app/rootfs \
			-v $(pwd)/config.json:/app/config.json \
			msc join 172.17.0.2:1234 && \
		sleep 10'" $OUTPUTNAME.cast && \
	docker run --rm -v $PWD:/data asciinema/asciicast2gif $OUTPUTNAME.cast $OUTPUTNAME.gif
tmux kill-session -t msc-asciinema
