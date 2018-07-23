GOOS=linux GOARCH=amd64 go build -o ./build-quiz-bot github.com/stels-cs/bot-clicker
ssh dp@web4.vkforms.ru 'rm -rf /home/dp/clicker-bot/build-quiz-bot.old.2'
ssh dp@web4.vkforms.ru 'mv /home/dp/clicker-bot/build-quiz-bot.old /home/dp/clicker-bot/build-quiz-bot.old.2'
ssh dp@web4.vkforms.ru 'mv /home/dp/clicker-bot/build-quiz-bot /home/dp/clicker-bot/build-quiz-bot.old'
scp build-quiz-bot dp@web4.vkforms.ru:/home/dp/quiz-bot/
ssh dp@web4.vkforms.ru 'killall -s SIGUSR1 build-quiz-bot'
sleep 1
ssh dp@web4.vkforms.ru 'ps aux | grep build-quiz-bo[t]'
sleep 1
ssh dp@web4.vkforms.ru 'tail -100 /home/dp/quiz-bot/log.log'
ssh dp@web4.vkforms.ru 'date'