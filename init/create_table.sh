#!/bin/sh

CMD_MYSQL="mysql -u${MYSQL_USER} -p${MYSQL_PWD} ${MYSQL_DATABASE}"
$CMD_MYSQL -e "create table hackathon (
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
