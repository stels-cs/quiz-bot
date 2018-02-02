package Vk

// see https://vk.com/dev/errors
const ApiUnknowmError = 1 //Произошла неизвестная ошибка.
const ApiError2 = 2 //Приложение выключено.
const ApiError3 = 3 //Передан неизвестный метод.
const ApiError4 = 4 //Неверная подпись.
const ApiAuthError = 5 //Авторизация пользователя не удалась.
const ApiTooManyRequests = 6 //Слишком много запросов в секунду.
const ApiError7 = 7 //Нет прав для выполнения этого действия.
const ApiError8 = 8 //Неверный запрос.
const ApiTooManyActions = 9 //Слишком много однотипных действий.
const ApiSearverError = 10 //Произошла внутренняя ошибка сервера.
const ApiError11 = 11 //В тестовом режиме приложение должно быть выключено или пользователь должен быть залогинен.
const ApiErrorCaptcha = 14 //Требуется ввод кода с картинки (Captcha).
const ApiError15 = 15 //Доступ запрещён.
const ApiError16 = 16 //Требуется выполнение запросов по протоколу HTTPS, т.к. пользователь включил настройку, требующую работу через безопасное соединение.
const ApiError17 = 17 //Требуется валидация пользователя.
const ApiError18 = 18 //Страница удалена или заблокирована.
const ApiError20 = 20 //Данное действие запрещено для не Standalone приложений.
const ApiError21 = 21 //Данное действие разрешено только для Standalone и Open API приложений.
const ApiError23 = 23 //Метод был выключен.
const ApiError24 = 24 //Требуется подтверждение со стороны пользователя.
const ApiError27 = 27 //Ключ доступа сообщества недействителен.
const ApiError28 = 28 //Ключ доступа приложения недействителен.
const ApiError100 = 100 //Один из необходимых параметров был не передан или неверен.
const ApiError101 = 101 //Неверный API ID приложения.
const ApiError113 = 113 //Неверный идентификатор пользователя.
const ApiError150 = 150 //Неверный timestamp.
const ApiError200 = 200 //Доступ к альбому запрещён.
const ApiError201 = 201 //Доступ к аудио запрещён.
const ApiError203 = 203 //Доступ к группе запрещён.
const ApiError300 = 300 //Альбом переполнен.
const ApiError500 = 500 //Действие запрещено. Вы должны включить переводы голосов в настройках приложения.
const ApiError600 = 600 //Нет прав на выполнение данных операций с рекламным кабинетом.
const ApiError603 = 603 //Произошла ошибка при работе с рекламным кабинетом