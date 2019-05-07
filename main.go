package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

// ClientMap ...

/* type DbWorker struct {
	Dsn string
	Db  *sql.DB
} */

// MyData struct function is to store my data. For example many of net.Conn and database
type MyData struct {
	conn      net.Conn    //self
	clients   *[]net.Conn //record online client
	db        *sql.DB
	clientmap map[string]net.Conn //record key of usename and value of tcp link
	username  string              //record self username
	password  string              //record self password
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

// CreateNewRoom function to create a new room to chat
func CreateNewRoom(linkdata *MyData) {
	for {
		str := "Welcome to create a new Room for chat,please input new Room num，must be digital\n"
		linkdata.conn.Write([]byte(str))
		roomnum, err1 := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err1 != nil {
			fmt.Println("CreateNewRoom function's bufio find error,the function will go done")
			break
		}

		length := len(roomnum)
		roomnum = roomnum[0 : length-1]

		//var tablename string
		err := linkdata.db.QueryRow("select * from chatroom.groups where group_id=?", roomnum).Scan()
		if err == sql.ErrNoRows {
			linkdata.db.QueryRow("insert into chatroom.groups values(?,?)", roomnum, linkdata.username)
			str := "please invite some friend to your room.between user and user use ',' ,thanks\n"
			str1 := AllRegisteredUser(linkdata)
			str += str1
			str2 := fmt.Sprintf("(your username is %s)\n\r\n", linkdata.username)
			str += str2
			linkdata.conn.Write([]byte(str))
			//this is all message about into new create room's people
			friends, err2 := bufio.NewReader(linkdata.conn).ReadString('\r')
			if err2 != nil {
				fmt.Println("CreateNewRoom function's bufio find error,the function will go done")
				break
			}

			f := func(c rune) bool {
				return !unicode.IsLetter(c) && !unicode.IsNumber(c)
			}
			alluser := strings.FieldsFunc(friends, f)
			for _, str3 := range alluser {
				linkdata.db.QueryRow("insert into chatroom.group_users values(?,?)", str3, roomnum)
			}
			linkdata.db.QueryRow("insert into chatroom.group_users values(?,?)", linkdata.username, roomnum)
			strg := "Ok! The friend already add to the room\r\n"
			linkdata.conn.Write([]byte(strg))
			break
		} else {
			str := "sorry! the room num existence,please press again\n"
			linkdata.conn.Write([]byte(str))
		}
	}
}

// AllRegisteredUser function can return all user who are already registered
func AllRegisteredUser(linkdata *MyData) string {
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
	return str
}

// ChatOnePerson function is one to one to talk
func ChatOnePerson(linkdata *MyData) {
	for {

		str := AllRegisteredUser(linkdata)
		str += "please select a person.between username and message use Space,thanks\n"
		linkdata.conn.Write([]byte(str))
		/*
			1：先检查输入的用户是不是存在
			2：存在的话，看看这个用户在没在线
			3：在线的话就直接转发
			4：不在线的话就把这个消息存储在离线表里面
		*/
		mess, err2 := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err2 != nil {

			fmt.Println("ChatOnePerson function.sorry,we don't get nothing in I/O")
			break
		}
		num := strings.Index(mess, " ")
		username := mess[0:num]
		data := mess[num+1:]
		fmt.Println("username ", "data ", username, data)

		var password string
		var status int
		err1 := linkdata.db.QueryRow("select *from chatroom.users where username=?", username).Scan(&username, &password, &status)
		if err1 == sql.ErrNoRows {
			str := "this user does not exist,please press again it\n"
			linkdata.conn.Write([]byte(str))
			fmt.Println("this user does not exist")
			continue
		}
		//然后查看这个对应的用户名的连接是不是在线的，如果在线的话就直接转发，不在线的发就存储在离线的数据库里面

		//等于1 就直接发送
		if status == 1 {
			friend := linkdata.clientmap[username]
			str := fmt.Sprintf("%s :", linkdata.username)
			str = str + data + "\n"

			friend.Write([]byte(str))

			fmt.Println("send message success")
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

// ShowSelfRoomnum function can show all room
func ShowSelfRoomnum(linkdata *MyData) {
	var allnum []int
	rows, err := linkdata.db.Query("SELECT * FROM chatroom.group_users WHERE username = ?", linkdata.username)
	if err != nil {
		fmt.Println("query error,maybe the database already no link")
	}
	if rows.Next() {
		var roomnum int
		var username string
		err := rows.Scan(&username, &roomnum)
		if err != nil {
			fmt.Println("scan user failed")
		}
		allnum = append(allnum, roomnum)
	}
	if len(allnum) == 0 {
		str := "sorry! you don't join every room\n"
		linkdata.conn.Write([]byte(str))
	} else {
		str := "All roomnum: "
		for i := 0; i < len(allnum); i++ {
			str += strconv.Itoa(allnum[i]) + " "
		}
		str += "\n"
		linkdata.conn.Write([]byte(str))
	}
}

// QuitOneRoomnum quit a room or admin to delete all room
func QuitOneRoomnum(linkdata *MyData) {
	//首先查看这个用户是不是admin 是的话就所有都删除，不是的话就只让这个用户从所属的表里面退出
	flag := 3
	for {
		ShowSelfRoomnum(linkdata)
		str := "please input diginal about your delete’s room\n"
		linkdata.conn.Write([]byte(str))
		str1, err1 := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err1 != nil {
			fmt.Println("QuitOneRoomnum function.sorry wo don't get nothing in I/O.")
			break
		}
		length := len(str1)
		str1 = str1[0 : length-1]
		num, _ := strconv.Atoi(str1)
		var groupid int
		var username string
		err := linkdata.db.QueryRow("SELECT * FROM chatroom.groups WHERE group_id= ?", num).Scan(&groupid, &username)
		if err == sql.ErrNoRows {
			str := "sorry.The roomnum don't exisence,please try again,thanks\n"
			linkdata.conn.Write([]byte(str))
			if flag == 0 {
				return
			}
			flag--
			continue
		} else {
			if username == linkdata.username {
				//如果相等的话，把groups和group_user里面的都删除了
				linkdata.db.Query("DELETE FROM chatroom.groups WHERE group_id =?", num)
				userpeople := AllRegisteredUser(linkdata)
				f := func(c rune) bool {
					return !unicode.IsLetter(c) && !unicode.IsNumber(c)
				}
				allusers := strings.FieldsFunc(userpeople, f)
				for _, val := range allusers {
					linkdata.db.Query("DELETE FROM chatroom.group_users WHERE username = ? AND group_id =?", val, num)
				}

			} else {
				//如果不相等的话，那么只删除在group_user里面的
				linkdata.db.Query("DELETE FROM chatroom.group_users WHERE username = ? AND group_id =?", linkdata.username, num)
			}
			break
		}
	}
}

// SelectRoom function can chat with mang people in room
func SelectRoom(linkdata *MyData) int {
	for {
		str := "please select a room to chat,thanks!\n"
		linkdata.conn.Write([]byte(str))
		ShowSelfRoomnum(linkdata)
		roomnum, err1 := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err1 != nil {
			fmt.Println("SelectRoom function.sorry. we don't get nothing in I/O")
			return -1
		}
		//先判断这个房间存在不存在

		err := linkdata.db.QueryRow("SELECT * FROM chatroom.groups WHERE group_id=?", roomnum).Scan()
		if err == sql.ErrNoRows {
			str := "sorry,the room don't exisence,please try again\n"
			linkdata.conn.Write([]byte(str))
			continue

		} else {
			for {
				str := "please input your message to others,use Enter end,thanks\n"
				linkdata.conn.Write([]byte(str))
				message, err1 := bufio.NewReader(linkdata.conn).ReadString('\r')
				if err1 != nil {
					fmt.Println("SelectRoom function.sorry. we don't get nothing in I/O")
					return -1
				}
				message = message[0 : len(message)-1]
				rows, err := linkdata.db.Query("SELECT * FROM chatroom.group_users WHERE group_id = ?", roomnum)
				if err != nil {
					fmt.Println("databases not link")
					return -1
				}

				for rows.Next() {
					var username string
					var id int
					err1 := rows.Scan(&username, &id)
					if err1 != sql.ErrNoRows {
						if username == linkdata.username {
							continue
						}
						other, ok := linkdata.clientmap[username]
						if ok {
							other.Write([]byte(message))
							other.Write([]byte("\n"))
						}
					}
				}
				str1 := "if you want to go away the interface,please input 'Y' or 'y',thanks\n"
				linkdata.conn.Write([]byte(str1))
				result, err2 := bufio.NewReader(linkdata.conn).ReadString('\r')
				if err2 != nil {
					fmt.Println("SelectRoom function.sorry. we don't get nothing in I/O")
					return -1
				}
				result = result[0 : len(result)-1]
				if result == "Y" || result == "y" {
					return 0
				}

			}
		}
	}
}

// SelectaRoom function select a room to chat
func SelectaRoom(linkdata *MyData) int {
	for {
		str := "1:SelectRoom\n"
		str += "2:exit\n"
		linkdata.conn.Write([]byte(str))
		num, _ := bufio.NewReader(linkdata.conn).ReadString('\r')
		length := len(num)
		num = num[0 : length-1]
		num1, _ := strconv.Atoi(num)
		switch num1 {
		case 1:
			SelectRoom(linkdata)
		case 2:
			return 2
		}
	}
}

// GotoChatroom function have two effect.first
func GotoChatroom(linkdata *MyData) int {
	for {
		str := "1.ShowSelfRoomnum\n"
		str += "2.QuitOneRoomnum\n"
		str += "3:SelectaRoom\n"
		str += "4:exit\n"
		linkdata.conn.Write([]byte(str))
		str1, _ := bufio.NewReader(linkdata.conn).ReadString('\r')
		length := len(str1)
		str1 = str1[0 : length-1]
		num, _ := strconv.Atoi(str1)
		fmt.Println("num ", num)
		switch num {
		case 1:
			ShowSelfRoomnum(linkdata)
		case 2:
			QuitOneRoomnum(linkdata)
		case 3:
			SelectaRoom((linkdata))
		case 4:
			return 4
		}
	}
}

//exit function have two effect.first set status to zero from chatromm.users.
func exit(linkdata *MyData) {
	var name string
	err := linkdata.db.QueryRow("update chatroom.users set status=0 where username=?", linkdata.username).Scan(&name)
	if err == sql.ErrNoRows {
		fmt.Println("users tables don't have the user")
	}
}

// NewMenu describe that we already go in deal with message part.
func NewMenu(linkdata *MyData) int {
	var mess string
	for {
		mess = "1:GetOnlinePerson\n"
		mess += "2:CreateNewRoom\n"
		mess += "3:ChatOnePerson\n"
		mess += "4:GotoChatroom\n"
		mess += "5:exit\n"
		linkdata.conn.Write([]byte(mess))

		data, err := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err != nil {
			fmt.Println("NewMenu function.sorry,we get nothing from i/o")
			return -1
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
			CreateNewRoom(linkdata)
		case 3:
			ChatOnePerson(linkdata)
		case 4:
			GotoChatroom(linkdata)
		case 5:
			exit(linkdata)
		}
		if num1 == 5 {
			return 5
		}
	}

}

// RegistAccount your count and passwd
func RegistAccount(linkdata *MyData) {
	for {
		str := "please register you count and password.\nbetween count and password use ','.\nuse 'Enter' end\n"
		linkdata.conn.Write([]byte(str))

		message, err1 := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err1 != nil {
			fmt.Println("LandingAccount function's bufio find error,the function will go done")
			break
		}

		fmt.Println("message", message)
		num := strings.Index(message, ",")
		count := message[0:num]
		fmt.Println(count)
		length := len(message)
		password := message[num+1 : length-1]
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
			linkdata.db.Query("insert into chatroom.users values(?,?,0)", count, password)
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
	flag := 0
	for {
		//先登陆（三次机会）
		str := "please press your count and passwordd\nbetween count and password use ','.\nuse 'Enter' end\n"
		linkdata.conn.Write([]byte(str))
		mess, err1 := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err1 != nil {
			fmt.Println("LandingAccount function's bufio find error,the function will go done")
			count := linkdata.username
			linkdata.db.QueryRow("update chatroom.users set status=0 where username=?", count).Scan()
			break
		}

		num := strings.Index(mess, ",")
		count := mess[0:num]
		fmt.Println("count ", count)

		passwor := mess[num+1:]
		password := passwor[0 : len(passwor)-1]
		fmt.Println("password ", password)
		var name string
		var code string
		var status int
		err := linkdata.db.QueryRow("select *from chatroom.users where username=? and userpassword=?", count, password).Scan(&name, &code, &status)
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
			//登陆成功以后看看有没有离线的消息
			str := "Congratulations!!! you are already Landing success.\n"
			linkdata.conn.Write([]byte(str))

			linkdata.db.QueryRow("update chatroom.users set status=1 where username=?", count).Scan(&name)
			linkdata.clientmap[count] = linkdata.conn
			linkdata.username = count
			linkdata.password = password
			offlinedata := make(map[string]string)
			//查找offline表，是不是有离线的消息
			var message string
			err = linkdata.db.QueryRow("select *from chatroom.offlineuser where username=?", count).Scan(&name, &message)
			if err == sql.ErrNoRows {
				fmt.Println("The user don't have offline message")

			}
			if message != "" {
				offlinedata["offline"] = message
				str := "you have a offline message,Do you want to read?\npelase press yes and no \n"
				linkdata.conn.Write([]byte(str))
				//to recv from client
				result, _ := bufio.NewReader(linkdata.conn).ReadString('\r')
				if result == "yes\r" {
					linkdata.conn.Write([]byte(offlinedata["offline"]))
					linkdata.conn.Write([]byte("\n"))
				}
				linkdata.db.QueryRow("update chatroom.offlineuser set message='' where username=?", count).Scan(&name)
			}
			num := NewMenu(linkdata)
			if num == 5 {
				break
			}
		}
	}

}

// ShowMenu show menu when a client logon.
func ShowMenu(linkdata *MyData) error {
	for {
		var mess string
		mess = "1:registered\n"
		mess += "2:Landing\n"

		linkdata.conn.Write([]byte(mess))

		message, err := bufio.NewReader(linkdata.conn).ReadString('\r')
		if err != nil {
			return fmt.Errorf("read from conn faild: %v", err)
		}

		fmt.Println("message", message)
		length := len(message)
		str := message[0 : length-1]

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
	clientmap := make(map[string]net.Conn)
	//num := 0
	// accept connection on port
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/chatroom")
	defer db.Close()

	err1 := db.Ping()
	if err1 != nil {
		fmt.Println(err1)
		fmt.Println("mydata link error,process will go done,thanks")
	}

	//mutex = &sync.MUTEX{}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept failed: %v", err)
			continue
		}
		defer conn.Close()

		clients = append(clients, conn)

		linkdata := MyData{conn, &clients, db, clientmap, "", ""}

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
