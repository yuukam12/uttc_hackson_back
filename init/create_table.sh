#!/bin/sh

CMD_MYSQL="mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}"
$CMD_MYSQL -e "create table content (
    id int(10) AUTO_INCREMENT NOT NULL primary key,
    title varchar(50) NOT NULL,
    description varchar(1000),
    url varchar(255),
    image LONGBLOB,
    uploaded_by varchar(50),
    create_data datetime
    category varchar(50)
    media varchar(50)
    );"

$CMD_MYSQL -e  "insert into content values (1, '記事1', '記事1です。');"