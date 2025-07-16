.PHONY: run


run:
	@while true; do \
		DEBUG=true go run ./cmd/main.go -process-tag "debug-lazy-apple-music-instance"; \
		sleep 0.2; \
	done

autoload:
	@fswatch -r -E ".*\.go$$" ./cmd ./internal | while read file; do \
		sleep 0.2; \
		echo "restarting lazyAppleMusic..."; \
		pkill -f 'main -process-tag debug-lazy-apple-music-instance' || true; \
	done

tmux:
	@tmux kill-session -t lazyAppleMusic || true && \
	tmux new-session -d -s lazyAppleMusic -n main && \
	\
	tmux send-keys -t lazyAppleMusic:main.1 "make run" C-m && \
	tmux split-window -h -t lazyAppleMusic:main.1 && \
	\
	tmux send-keys -t lazyAppleMusic:main.2 "tail -f tmp/debug.log" C-m && \
	tmux split-window -v -t lazyAppleMusic:main.2 && \
	\
	tmux send-keys -t lazyAppleMusic:main.3 "make autoload" C-m && \
	tmux split-window -v -t lazyAppleMusic:main.3 && \
	\
	tmux send-keys -t lazyAppleMusic:main.4 && \
	tmux select-layout -t lazyAppleMusic:main && \
	tmux attach -t lazyAppleMusic
