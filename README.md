# Go实现的聊天室程序

## 需要的软件支持或者需要了解什么

1：数据库(我是用的是MySQL Workbench 8.0 CE）
2：熟练使用Go语言
3：熟悉Go语言的TCP通信方式
4：vs code的熟练使用

## 表结构的设计

### users表

主键是username(用户名)，还有userpassword(用户密码)，还有一个flag(用户在没在线)

### offlineuser表

主键是username(用户名)，还有一个message(存储用户的)

### groups表

这个表的主键是group_id也就是房间号码，第二个是username(主要存储的就是到底是谁创建了这个房间，也就是群主)

### group_users表

这个表有两个值。一个是username，另一个是group_id。主键是由他们两个一块组成的，是联合主键

## 完成的功能

1：注册模块
2：登陆模块
3：展示在线的人数
4：一对一通信
5：多对多通信

### 注册模块

用户输入账户(主键)和密码，然后存储在后端的users表里面

### 登陆模块

用户输入账户和密码，然后再后端的users表里面去查找，有的话就告诉用户登陆成功，没有的话，让用户重新去输入，当输入的次数达到3次上限的话，就退出到，注册和登陆的界面。

### 展现在线的人数

再users表里面返回flag为1的用户名

### 一对一通信

首先第一步返回再users表里面注册的人数，返回给用户。当服务器这表读到了用户输入的username+message然后先查找这个用户在没在users表里面，如果在的话就直接发送给这个用户，如果不在的话，就发送到离线表里面，当这个用户在此上线的话。离线表里面的信息就会发送过来

### 多对多通信

首先在我们选择进入多用户聊天的时候，会出现让你输入你的房间号，假如你输入的房间号是错误的，而且累加到了三次，就从这一层退出了。那假如你输入的是正确的，就会让你输入信息，然后就会传送到在线的玩家的终端上面，这样就实现了多人聊天室的通信。
而且对于删除房间的操作也有两步，第一步当时admin的时候的做法时，删除在这个房间里面的所有的人，第二步当不是admin的时候的做法就是只是自己从那个群里面退出来了而已。