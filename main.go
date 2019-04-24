package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// ClientMap ...

/* type DbWorker struct {
	Dsn string
	Db  *sql.DB
} */

// MyData struct function is to store my data. For example many of net.Conn and database
type MyData struct {
	conn      net.Conn
	clients   *[]net.Conn
	db        *sql.DB
	clientmap *map[string]net.Conn
}

// BroadcastMessage broadcast received message to all clients currently connected.
func BroadcastMessage(conn net.Conn, clients []net.Conn) {
	for {
		// will listen for message to process ending in newline (\n)
		fmt.Println("BroadcastMessage func start")
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// send new string back to client
		for _, cli := range clients {
			cli.Write([]byte(message))
		}
	}
}

// GetOnlinePerson function is to return that status is '1' object's name
func GetOnlinePerson(linkdata *MyData) {
	var str string
	rows, err := linkdata.db.Query("select *from chatroom.users where status=1")
	if err != nil {
		fmt.Println("Query data error")
	}
	fmt.Println("Query yes")
	for rows.Next() {
		fmt.Println("rows start")
		var username string
		var userpassword string
		var status int
		err = rows.Scan(&username, &userpassword, &status)
		if err != nil {
			fmt.Println("Query data error1")
		}
		str += username + "\n"
	}
	linkdata.conn.Write([]byte(str))
}

// ChatOnePerson function is one to one to talk
func ChatOnePerson(linkdata *MyData) {
	for {
		var str string
		str = "This is all already registered person\n"
		rows, err := linkdata.db.Query("select *from chatroom.users")
		if err != nil {
			fmt.Println("Query data error")
		}
		for rows.Next() {
			var username string
			var userpassword string
			var status int
			err := rows.Scan(&username, &userpassword, &status)
			if err != nil {
				fmt.Println("Query data error1")
			}
			str += username + "\n"
		}
		str += "please select a person.between username and message use Space,thanks\n"
		linkdata.conn.Write([]byte(str))
		/*
			1：先检查输入的用户是不是存在
			2：存在的话，看看这个用户在没在线
			3：在线的话就直接转发
			4：不在线的话就把这个消息存储在离线表里面
		*/
		mess, err := bufio.NewReader(linkdata.conn).ReadString('\r')
		num := strings.Index(mess, " ")
		username := mess[0:num]
		data := mess[num+1:]
		fmt.Println("username ", "data ", username, data)

		var password string
		var status int
		err1 := linkdata.db.QueryRow("select *from chatroom.users where username=?", username).Scan(&username, &password)
		if err1 == sql.ErrNoRows {
			str := "this user does not exist,please press again it\n"
			linkdata.conn.Write([]byte(str))
			fmt.Println("this user does not exist")
			continue
		}
		//然后查看这个对应的用户名的连接是不是在线的，如果在线的话就直接转发，不在线的发就存储在离线的数据库里面

		//等于1 就直接发送
		if status == 1 {
			friend := (*(linkdata.clientmap))[username]
			friend.Write([]byte(data))
			break
		} else {
			//发送信息到离线表里面
			var str string
			var name string
			err := linkdata.db.QueryRow("update chatroom.offlineuser set message=? where username=?", data, username).Scan(&name, &str)
			if err != sql.ErrNoRows {
				fmt.Println("update data error")
			}
			break
		}

	}

}

// NewMenu describe that we already go in deal with message part.
func NewMenu(linkdata *MyData) {
	var mess string
	for {
		mess = "1:GetOnlinePerson\n"
		mess += "2:CreateNewRoom\n"
		mess += "3:ChatOnePerson\n"
		mess += "4:SelectOnlineRoom\n"
		linkdata.conn.Write([]byte(mess))

		data, err := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err != nil {
			fmt.Println("sorry,we get nothing from i/o")
		}
		num := strings.Index(data, "\r")
		num1, err := strconv.Atoi(data[0:num])
		if err != nil {
			fmt.Println("type convert error")
			continue
		}
		fmt.Println("num1", num1)
		switch num1 {
		case 1:
			GetOnlinePerson(linkdata)
		case 2:
			//CreateNewRoom()
		case 3:
			ChatOnePerson(linkdata)
		case 4:
			//SelectOnlineRoom()

		}

	}

}

// RegistAccount your count and passwd
func RegistAccount(linkdata *MyData) {
	for {
		str := "please register you count and password.\nbetween count and password use ','.\nuse 'Enter' end\n"
		linkdata.conn.Write([]byte(str))

		message, _ := bufio.NewReader(linkdata.conn).ReadString('\r')
		fmt.Println("message", message)
		num := strings.Index(message, ",")
		count := message[0:num]
		fmt.Println(count)
		password := message[num+1:]
		fmt.Println(password)
		/*
			在数据库进行检测
			1：如果在users表里面的话，就进入到UserMenu()这个新的界面
			2：如果没有在users这个表里面的话，就打印出错误，重新输入即可
		*/
		/* 	linkdata.db.Query("use chatroom.users")
		stmt, err := linkdata.db.Query("select *from chatroom.users where username='juwenjie'")
		defer stmt.Close() */
		var name string
		err := linkdata.db.QueryRow("select *from chatroom.users where username =?", count).Scan(&name)
		//rows, err := linkdata.db.Query(0)
		//	defer rows.Close()
		if err == sql.ErrNoRows {
			fmt.Println("no find it, new count will create,thanks")
			linkdata.db.Query("insert into chatroom.users values(?,?)", count, password)
			linkdata.db.Query("insert into chatroom.offlineuser values(?,?)", count, "")
			str1 := "RegistCount success,thanks\r\n"
			linkdata.conn.Write([]byte(str1))
			break
		} else {
			fmt.Println("count exist")
			str1 := "count exist,please press new count\n\r\n"
			linkdata.conn.Write([]byte(str1))
			//fmt.Println("insert error", stmt)
			continue
		}
	}
}

// LandingAccount your count and password
func LandingAccount(linkdata *MyData) {
	var flag int = 0
	for {
		str := "please press your count and passwordd\nbetween count and password use ','.\nuse 'Enter' end\n"
		linkdata.conn.Write([]byte(str))
		mess, _ := bufio.NewReader(linkdata.conn).ReadString('\r')
		num := strings.Index(mess, ",")
		count := mess[0:num]
		fmt.Println("count ", count)
		password := mess[num+1:]
		fmt.Println("password ", password)
		var name string
		var code string
		err := linkdata.db.QueryRow("select *from chatroom.users where username=? and userpassword=?", count, password).Scan(&name, &code)
		if err == sql.ErrNoRows { //no find the message in database
			flag++
			if flag == 3 {
				str := "sorry!,your are already to press Upper limit.Bye Bye!\n"
				linkdata.conn.Write([]byte(str))
				break
			}
			str1 := "sorry,your count or password have some error,please after confirm to press again,thanks\n\r\n"
			linkdata.conn.Write([]byte(str1))
		} else {
			str := "Congratulations!!! you are already Landing success.\n"
			linkdata.conn.Write([]byte(str))
			linkdata.db.QueryRow("update chatroom.users set status=1 where username=?", count).Scan(&name)
			(*(linkdata.clientmap))[count] = linkdata.conn

			NewMenu(linkdata)
		}
	}

}

// ShowMenu show menu when a client logon.
func ShowMenu(linkdata *MyData) int {

	for {
		var mess string
		mess = "1:registered\n"
		mess += "2:Landing\n"

		linkdata.conn.Write([]byte(mess))

		message, _ := bufio.NewReader(linkdata.conn).ReadString('\r')
		fmt.Println("message", message)
		num := strings.Index(message, "\r")
		fmt.Println(num)
		str := message[0:num]

		num1, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("string conver int error")
			fmt.Println(err)
		}
		fmt.Println("num value", num1)
		switch num1 {
		case 1:
			RegistAccount(linkdata)
		case 2:
			LandingAccount(linkdata)
		}
		//conn.Write([]byte(mess))
		//return 1
	}
}

func main() {
	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Printf("listen failed: %v", err)
		os.Exit(1)
	}
	defer ln.Close()
	var clients []net.Conn
	//num := 0
	// accept connection on port
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/chatroom")
	defer db.Close()
	//mutex = &sync.MUTEX{}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept failed: %v", err)
			continue
		}
		defer conn.Close()

		clients = append(clients, conn)
		clientmap := make(map[string]net.Conn)

		linkdata := MyData{conn, &clients, db, &clientmap}

		go func() {
			ShowMenu(&linkdata)
		}()

		/*
			if err != nil {
				fmt.Println("mysql use error")
				panic(err)
			}
			defer db.Close()
			db.Query("use chatroom.users")
			fmt.Println("db value", db, "err value", err)
			db.Query("insert into chatroom.users values(101, 'liujili')")
			query, err := db.Query("select * from chatroom.users")

			v := reflect.ValueOf(query)
			fmt.Println(v)
			db.Query("insert into chatroom.users values(188, 'lili')") */
		//query, err := db.Query("select * from tmpdb.tmptab")
		//stmt, err := db.Prepare(`INSERT INTO users(username,userpassword) VALUES (132,'juwenjie')`)
		//fmt.Println(stmt, err)
		/*
			go func() {
				ShowMenu(conn, clients)
				//BroadcastMessage(conn, clients)
			}() */
	}
}
