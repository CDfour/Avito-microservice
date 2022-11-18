# Avito-microservice
Микросервис для работы с балансом пользователей

Микросервис осуществляет зачисление средств пользователям, перевод средств от пользователя к пользователю, заказ услуг, а также предоставляет информацию о балансе
клиента, месячные отчеты по всем пользователям для каждой из услуг и индивидуальную историю операций  

Инструкция по запуску
---------------------

Для запуска требуется ввести команду ```docker compose up``` в папке проекта  
Папкой проекта является директория путь к которой прописан в инструкции ```WORKDIR``` докерфайла, нужно поменять этот путь в инструкции ```WORKDIR``` если проект  
лежит в другой директории  

БД
---------

В папке проекта лежит файл ```init.sql``` с созданием всех необходимых таблиц в БД

Описание эндпоинтов
---------------------

После запуска проекта просмотр swagger-документации возможен по ссылке http://localhost:9000/swagger/index.html  

http://localhost:9000/balance [get]:  
Принимает id пользователя из параметров строки и возвращает JSON с информацией о данном пользователе    

http://localhost:9000/balance [post]:  
Принимает JSON вида:  
```{```  
```"id: <uuid пользователя>,```  
```"funds": <кол-во денег для пополнения баланса>,```  
```}```  
Пополняет баланс данного пользователя этим кол-вом средств  
Если пользователя нет в базе данных то он вносится в нее с данным балансом    

http://localhost:9000/transfer [post]:  
Принимает JSON вида:  
```{```  
```"sender_id": <uuid отправителя>,```  
```"recipient_id": <uuid получателя>,```  
```"funds": <кол-во денег для перевода>,```  
```}```  
Переводит эти средства от одного пользователя к другому  

http://localhost:9000/order [post]:  
Принимает JSON вида:  
```{```  
```"user_id": <uuid пользователя>,```  
```"service_id": <uuid услуги>,```  
```"service_name": <"Название услуги">,```  
```"cost": <стоимость услуги>,```  
```}```  
Осуществляет резервацию денег  
Название услуги добавлено в запрос для отчета в котором данные собираются по всем пользователям сгруппированые по названиям услуг  

http://localhost:9000/order/success [post]:  
Принимает JSON вида:  
```{```  
```"user_id": <uuid пользователя>,```  
```"service_id": <uuid услуги>,```  
```"service_name": <"Название услуги">,```  
```"cost": <стоимость услуги>,```  
```}```  
Осуществляет разрезервирование выполненного заказа    

http://localhost:9000/order/failed [post]:  
Принимает JSON вида:  
```{```  
```"user_id": <uuid пользователя>,```  
```"service_id": <uuid услуги>,```  
```"service_name": <"Название услуги">,```  
```"cost": <стоимость услуги>,```  
```}```  
Осуществляет разрезервирование невыполненного заказа   

http://localhost:9000/report [post]:  
Принимает JSON вида:  
```{```  
```"year": <"год">,```  
```"month": <"месяц">,```  
```}```  
Cоздает месячный отчет по всем пользователям сгруппированный по названиям услуг и возвращает ссылку, с включенным в нее id (именем файла), по которому возможен просмотр отчета  
Файл формата ```.csv``` с соответствующим именем создается в папке reports    

http://localhost:9000/report/csv [get]:  
Принимает id файла из параметров строки и выводит отчет, названный этим id    

http://localhost:9000/history [get]:  
Принимает из параметров строки id пользователя и параметры ```limit,offset```  
Возвращает историю операций данного пользователя  




