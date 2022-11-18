# Avito-microservice
Микросервис для работы с балансом пользователей

Микросервис осуществляет зачисление средств пользователям, перевод средств от пользователя к пользователю, заказ услуг, а также предоставляет информацию о балансе клиента, месячные отчеты по всем пользователям для каждой из услуг и индивидуальную историю операций

Инструкция по запуску
---------------------

Для запуска требуется ввести команду docker compose up в папке проекта  
Папкой проекта является директория путь к которой прописан в инструкции WORKDIR, нужно поменять этот путь в инструкции WORKDIR если проект лежит в другой директории

БД
---------

В папке проекта лежит файл init.sql с созданием всех необходимых таблиц в БД

Описание эндпоинтов
---------------------

После запуска проекта просмотр swagger-документации возможен по ссылке http://localhost:9000/swagger/index.html  

http://localhost:9000/balance [get]:  
Принимает id пользователя из параметров строки и возвращает JSON с информацией о данном пользователе  
http://localhost:9000/balance [post]:  
Принимает JSON с id пользователя и кол-вом средств и пополняет баланс данного пользователя этим кол-вом средств  
                Если пользователя нет в базе данных то он вносится в нее с данным балансом  
http://localhost:9000/transfer [post]:  
Принимает JSON с id отправителя, id получателя и кол-во средств, переводит эти средства от одного к другому  
http://localhost:9000/order [post]:  
Принимает JSON с id пользователя, id услуги, название услуги, id заказа, стоимость и осуществляет резервацию  
http://localhost:9000/order/success [post]:  
Принимает JSON с id пользователя, id услуги, название услуги, id заказа, стоимость и осуществляет разрезервирование выполненного заказа  
http://localhost:9000/order/failed [post]:  
Принимает JSON с id пользователя, id услуги, название услуги, id заказа, стоимость и осуществляет разрезервирование невыполненного заказа  
http://localhost:9000/report [post]:  
Принимает JSON с годом и месяцем, создает отчет по этому месяцу и возвращает ссылку, с включенным в нее id файла, по которой возможен просмотр отчета  
http://localhost:9000/report/csv [get]:  
Принимает id файла из параметров строки и выводит соответствующий отчет  
http://localhost:9000/history [get]:  
Принимает из параметров строки id пользователя и параметры limit,offset. Возвращает историю операций данного пользователя  




