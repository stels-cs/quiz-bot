GOOS=linux GOARCH=amd64 go build -o ./build-quiz-bot github.com/stels-cs/quiz-bot
ssh dp@prod 'rm -rf /bot/build-quiz-bot.old.2'
ssh dp@prod 'mv /bot/build-quiz-bot.old /bot/build-quiz-bot.old.2'
ssh dp@prod 'mv /bot/build-quiz-bot /bot/build-quiz-bot.old'
scp build-quiz-bot dp@prod:/bot/
ssh dp@prod 'killall -s SIGUSR1 build-quiz-bot'
sleep 1
ssh dp@prod 'ps aux | grep build-quiz-bo[t]'
sleep 1
ssh dp@prod 'tail -100 /bot/log.log'
ssh dp@prod 'date'
