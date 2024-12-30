package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"middleproject/scripts"

	_ "github.com/go-sql-driver/mysql"
)

func AdmLogin(c *gin.Context) {
	var requestData model.LoginRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	var storedPassword string
	var userID string
	var userName string
	var Avatar string
	var peimission int
	isEmail := isEmailFormat(requestData.Userid)
	var query string
	if isEmail {
		query = "SELECT user_id, password, Uname, avatar, peimission FROM users WHERE email = ?"
	} else {
		query = "SELECT user_id, password, Uname, avatar, peimission FROM users WHERE user_id = ?"
	}
	row := db.QueryRow(query, requestData.Userid)
	info := row.Scan(&userID, &storedPassword, &userName, &Avatar, &peimission)

	if info != nil {
		if info == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"isok": false, "failreason": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库查询失败"})
		return
	}
	if storedPassword != requestData.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "密码错误"})
		return
	}
	if peimission != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "您不是管理员，请使用客户端登录"})
		return
	}
	err, Avatar = scripts.GetUrl(Avatar)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": Avatar})
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "uid": userID, "uname": userName, "uimage": Avatar})
}

type Userinfo struct {
	Uid    string `json:"uid"`
	Uimage string `json:"uimage"`
	Uname  string `json:"uname"`
}

func GetallUser(c *gin.Context) {
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	var users []Userinfo
	if err != nil || page == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"datas": users, "totalPages": 0})
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
		return
	}
	defer db.Close()
	query := "SELECT user_id, Uname, avatar FROM users limit ?, 10"
	rows, err := db.Query(query, page*10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
		return
	}
	for rows.Next() {
		var user Userinfo
		err = rows.Scan(&user.Uid, &user.Uname, &user.Uimage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		err, user.Uimage = scripts.GetUrl(user.Uimage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		users = append(users, user)
	}
	query = "SELECT count(*) FROM users"
	row := db.QueryRow(query)
	var total int
	err = row.Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
		return
	}
	totalPages := total / 10
	if total%10 != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, gin.H{"datas": users, "totalPages": totalPages})
}

type Postinfo struct {
	Postid      string   `json:"id"`
	Posttitle   string   `json:"title"`
	Uid         string   `json:"uid"`
	Uname       string   `json:"uname"`
	Uimage      string   `json:"uimage"`
	Time        string   `json:"time"`
	Somecontent string   `json:"content"`
	Subjects    []string `json:"subjects"`
}

func GetallPost(c *gin.Context) {
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	var posts []Postinfo
	if err != nil || page == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"logs": posts, "totalPages": 0})
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	defer db.Close()
	query := "select post_id,posts.user_id,Uname,avatar,title,content,post_subject,publish_time from posts,users where posts.user_id = users.user_id limit ?, 10"
	rows, err := db.Query(query, page*10)
	fmt.Println("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	for rows.Next() {
		var post Postinfo
		var subjects sql.NullString
		err = rows.Scan(&post.Postid, &post.Uid, &post.Uname, &post.Uimage, &post.Posttitle, &post.Somecontent, &subjects, &post.Time)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		if subjects.Valid {
			str := subjects.String
			post.Subjects = strings.Split(str[1:len(str)-1], ",")
			//去除双引号
			for i := 0; i < len(post.Subjects); i++ {
				if i == 0 {
					post.Subjects[i] = "#" + post.Subjects[i][1:len(post.Subjects[i])-1]

				} else {
					post.Subjects[i] = "#" + post.Subjects[i][2:len(post.Subjects[i])-1]
				}
			}

		}
		if len(post.Somecontent) > 300 {
			post.Somecontent = post.Somecontent[:300] + "..."
		}
		post.Time = post.Time[:len(post.Time)-3]
		err, post.Uimage = scripts.GetUrl(post.Uimage)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		posts = append(posts, post)
	}
	query = "SELECT count(*) FROM posts"
	row := db.QueryRow(query)
	var total int
	err = row.Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	totalPages := total / 10
	if total%10 != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, gin.H{"logs": posts, "totalPages": totalPages})

}

func AdmSearchUser(c *gin.Context) {
	aimUidstr := c.DefaultQuery("aimuid", "-1")
	aimUname := c.DefaultQuery("aimuname", "")
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	uid, err_str2int := strconv.Atoi(aimUidstr)
	type Userinfo struct {
		Uid    string `json:"uid"`
		Uimage string `json:"uimage"`
		Uname  string `json:"uname"`
	}
	var users []Userinfo
	var totalPage int
	if err != nil || page == -1 || err_str2int != nil {
		c.JSON(http.StatusBadRequest, gin.H{"datas": users, "totalPages": 0})
	}
	page = page - 1
	if uid == -1 {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "SELECT user_id, Uname, avatar FROM users WHERE Uname like ? limit ?, 10"
		rows, err := db.Query(query, "%"+aimUname+"%", page*10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		for rows.Next() {
			var user Userinfo
			err = rows.Scan(&user.Uid, &user.Uname, &user.Uimage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
				return
			}
			err, user.Uimage = scripts.GetUrl(user.Uimage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
				return
			}
			users = append(users, user)
		}
		query = "SELECT count(*) FROM users WHERE Uname like ?"
		row := db.QueryRow(query, "%"+aimUname+"%")
		var temp int
		err = row.Scan(&temp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": users, "totalPages": 0})
			return
		}
		totalPage = temp / 10
		if temp%10 != 0 {
			totalPage++
		}
	} else {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "SELECT user_id, Uname, avatar FROM users WHERE user_id = ? limit ?, 10"
		rows, err := db.Query(query, uid, page*10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		for rows.Next() {
			var user Userinfo
			err = rows.Scan(&user.Uid, &user.Uname, &user.Uimage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
				return
			}
			err, user.Uimage = scripts.GetUrl(user.Uimage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
				return
			}

			users = append(users, user)
		}
		totalPage = 1
	}
	c.JSON(http.StatusOK, gin.H{"datas": users, "totalPages": totalPage})

}

func AdmSearchPost(c *gin.Context) {
	aimPostidstr := c.DefaultQuery("aimlogid", "-1")
	aimPosttitle := c.DefaultQuery("aimtitle", "")
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	postid, err_str2int := strconv.Atoi(aimPostidstr)
	type Postinfo struct {
		Postid    string   `json:"id"`
		Posttitle string   `json:"title"`
		Uid       string   `json:"uid"`
		Uname     string   `json:"uname"`
		Uimage    string   `json:"uimage"`
		Time      string   `json:"time"`
		Subjects  []string `json:"subjects"`
	}
	var posts []Postinfo
	var totalPage int
	if err != nil || page == -1 || err_str2int != nil {
		c.JSON(http.StatusBadRequest, gin.H{"logs": posts, "totalPages": 0})
	}
	page = page - 1
	if postid == -1 {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "select post_id,posts.user_id,Uname,avatar,title,post_subject,publish_time from posts,users where posts.user_id = users.user_id and title like ? limit ?, 10"
		rows, err := db.Query(query, "%"+aimPosttitle+"%", page*10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		for rows.Next() {
			var post Postinfo
			var subjects sql.NullString
			err = rows.Scan(&post.Postid, &post.Uid, &post.Uname, &post.Uimage, &post.Posttitle, &subjects, &post.Time)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
				return
			}
			if subjects.Valid {
				str := subjects.String
				post.Subjects = strings.Split(str[1:len(str)-1], ",")
				//去除双引号
				for i := 0; i < len(post.Subjects); i++ {
					if i == 0 {
						post.Subjects[i] = "#" + post.Subjects[i][1:len(post.Subjects[i])-1]
					} else {
						post.Subjects[i] = "#" + post.Subjects[i][2:len(post.Subjects[i])-1]
					}
				}
			}
			post.Time = post.Time[:len(post.Time)-3]
			err, post.Uimage = scripts.GetUrl(post.Uimage)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
				return
			}
			posts = append(posts, post)
		}
		query = "SELECT count(*) FROM posts WHERE title like ?"
		row := db.QueryRow(query, "%"+aimPosttitle+"%")
		var temp int
		err = row.Scan(&temp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		totalPage = temp / 10
		if temp%10 != 0 {
			totalPage++
		}
	} else {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "select post_id,posts.user_id,Uname,avatar,title,post_subject,publish_time from posts,users where posts.user_id = users.user_id and post_id = ?"
		rows, err := db.Query(query, postid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		for rows.Next() {
			var post Postinfo
			var subjects sql.NullString
			err = rows.Scan(&post.Postid, &post.Uid, &post.Uname, &post.Uimage, &post.Posttitle, &subjects, &post.Time)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
				return
			}
			if subjects.Valid {
				str := subjects.String
				post.Subjects = strings.Split(str[1:len(str)-1], ",")
				//去除双引号
				for i := 0; i < len(post.Subjects); i++ {
					if i == 0 {
						post.Subjects[i] = "#" + post.Subjects[i][1:len(post.Subjects[i])-1]
					} else {
						post.Subjects[i] = "#" + post.Subjects[i][2:len(post.Subjects[i])-1]
					}
				}
			}
			post.Time = post.Time[:len(post.Time)-3]
			err, post.Uimage = scripts.GetUrl(post.Uimage)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
				return
			}
			posts = append(posts, post)
		}
		totalPage = 1
	}
	c.JSON(http.StatusOK, gin.H{"logs": posts, "totalPages": totalPage})

}

// 生成封禁禁言警告系统消息(封禁、禁言、警告)，参数：类型，被举报类型（帖子、评论）、类型id(帖子id、评论id)，用户id，天数
func MakeSysinfo(Htype string, rtype string, id int, day int) (bool, string) {
	db, err := repository.Connect()
	if err != nil {
		return false, "数据库连接失败"
	}
	var content string
	if rtype == "log" {
		select_query := "SELECT title FROM posts WHERE post_id = ?"
		row := db.QueryRow(select_query, id)
		err := row.Scan(&content)
		if err != nil {
			return false, "查询帖子失败"
		}
		content = "您的帖子《" + content + "》违反社区规则，已被管理员删除。"

	} else if rtype == "comment" || rtype == "reply" {
		select_query := "SELECT content FROM comments WHERE comment_id = ?"
		row := db.QueryRow(select_query, id)
		err := row.Scan(&content)
		if err != nil {
			return false, "查询评论失败"
		}
		content = "您的评论《" + content + "》违反社区规则，已被管理员删除。"
	}
	var info string
	currentTime := time.Now()
	chinaTime := currentTime.Add(8 * time.Hour)
	if Htype == "封禁" {
		start := chinaTime
		end := start.Add(time.Duration(day) * 24 * time.Hour)
		var startstr string
		var endstr string
		startstr = start.Format("2006-01-02 15:04:05")
		endstr = end.Format("2006-01-02 15:04:05")
		info = "我们遗憾地通知您，由于您在本网站的行为违反了我们的社区规范，您的账户已被暂时封禁(" + startstr + "-" + endstr + ")。具体原因如下：\n  "
	} else if Htype == "禁言" {
		start := chinaTime
		end := start.Add(time.Duration(day) * 24 * time.Hour)
		var startstr string
		var endstr string
		startstr = start.Format("2006-01-02 15:04:05")
		endstr = end.Format("2006-01-02 15:04:05")
		info = "我们遗憾地通知您，由于您在本网站的行为违反了我们的社区规范，您的账户已被暂时禁言(" + startstr + "-" + endstr + ")。具体原因如下：\n  "
	} else if Htype == "警告" {
		info = "尊敬的用户，您好！您近期发布的内容因为违反社区规则，已被警告，希望您可以注意您的言行。"
		return true, info
	}
	info = info + content + "\n"
	info = info + "我们重视每一位用户的体验，并致力于维护一个健康、积极的社区环境。请您在未来遵守以下社区行为准则：\n"
	info = info + "1、尊重他人，保持友善的交流。\n2、禁止发布任何违反法律法规的内容。\n3、禁止发布任何侮辱、攻击、歧视性的言论。\n"
	return true, info
}

// 用户反馈(返回要存储在数据库的信息)，被处理人类型，天数和id
func UserFeedback(Htype string, day int, uid int) (bool, string) {
	db, err := repository.Connect()
	if err != nil {
		return false, "数据库连接失败"
	}
	defer db.Close()
	var uname string
	select_query := "SELECT Uname FROM users WHERE user_id = ?"
	row := db.QueryRow(select_query, uid)
	err = row.Scan(&uname)
	if err != nil {
		return false, "查询用户失败"
	}
	var infor string
	var content string
	infor = "尊敬的用户，您好！您向我们提出的反馈我们已经处理，处理结果如下：\n  "
	daystr := strconv.Itoa(day)
	if Htype == "封禁" {
		content = uname + "发布的内容因为违反社区规则，已被封禁" + daystr + "天。"
	} else if Htype == "禁言" {
		content = uname + "发布的内容因为违反社区规则，已被禁言" + daystr + "天。"
	} else if Htype == "警告" {
		content = uname + "发布的内容因为违反社区规则，已被警告。"
	}
	infor = infor + content + "\n"
	infor = infor + "  感谢您对净化社区环境的贡献，我们将继续努力，为您提供更好的服务！"
	return true, infor
}

// 内容删除通知
func ContentDelete(ContentType string, id int) (bool, string) {
	db, err := repository.Connect()
	if err != nil {
		return false, "数据库连接失败"
	}
	defer db.Close()
	var infor string
	if ContentType == "log" {
		var title string
		select_query := "SELECT title FROM posts WHERE post_id = ?"
		row := db.QueryRow(select_query, id)
		err = row.Scan(&title)
		if err != nil {
			return false, "查询帖子失败"
		}
		infor = "尊敬的用户，您好！您发布的帖子《" + title + "》因为违反社区规则，已被社区管理员删除。"
	} else if ContentType == "comment" || ContentType == "reply" {
		var content string
		select_query := "SELECT content FROM comments WHERE comment_id = ?"
		row := db.QueryRow(select_query, id)
		err = row.Scan(&content)
		if err != nil {
			return false, "查询评论失败"
		}
		infor = "尊敬的用户，您好！您发布的评论《" + content + "》因为违反社区规则，已被社区管理员删除。"
	}
	infor = infor + "创建绿色网络环境，还需我们共同努力！"
	return true, infor
}

// 管理员删除帖子
func AdmDeletePost(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	postidstr := c.DefaultQuery("logid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	postid, err_pid := strconv.Atoi(postidstr)
	if err_uid != nil || err_pid != nil || uid == -1 || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	query := "DELETE FROM posts WHERE post_id = ?"
	_, err = db.Exec(query, postid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子删除失败"})
		return
	}
	//将删除内容通知存储到系统消息表
	isok, info := ContentDelete("log", postid)
	if !isok {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": info})
		return
	}
	query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	infotype := "内容删除通知"
	_, err = db.Exec(query, uid, infotype, info)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})

}

// 管理员删除评论
func AdmDeleteComment(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	commentidstr := c.DefaultQuery("comid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	commentid, err_cid := strconv.Atoi(commentidstr)
	if err_uid != nil || err_cid != nil || uid == -1 || commentid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	querystr := "SELECT post_id FROM comments WHERE comment_id = ?"
	row := db.QueryRow(querystr, commentid)
	var postid int
	err = row.Scan(&postid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子id失败"})
		return
	}
	query := "DELETE FROM comments WHERE comment_id = ?"
	_, err = db.Exec(query, commentid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论删除失败"})
		return
	}
	//更新帖子评论数
	query = "UPDATE posts SET comment_count = comment_count - 1 WHERE post_id = ?"
	_, err = db.Exec(query, postid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子评论数更新失败"})
		return
	}

	isok, info := ContentDelete("comment", commentid)
	if !isok {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": info})
		return
	}
	query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	infotype := "内容删除通知"
	_, err = db.Exec(query, uid, infotype, info)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})
}

// 管理员删除回复
func AdmDeleteReply(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	replyidstr := c.DefaultQuery("replyid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	replyid, err_rid := strconv.Atoi(replyidstr)
	if err_uid != nil || err_rid != nil || uid == -1 || replyid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	//查询评论top_parentid
	query_2 := "SELECT top_parentid FROM comments WHERE comment_id = ?"
	row := db.QueryRow(query_2, replyid)
	var top_parentid int
	err = row.Scan(&top_parentid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询评论top_parentid失败"})
		return
	}
	query := "DELETE FROM comments WHERE comment_id = ?"
	_, err = db.Exec(query, replyid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "回复删除失败"})
		return
	}
	//更新评论数
	query = "UPDATE comments SET reply_count = reply_count - 1 WHERE comment_id = ?"
	_, err = db.Exec(query, top_parentid)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论回复数更新失败"})
		return
	}

	isok, info := ContentDelete("reply", replyid)
	if !isok {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": info})
		return
	}
	query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	infotype := "内容删除通知"
	_, err = db.Exec(query, uid, infotype, info)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})
}

// 管理员封禁与禁言
func AdmBan(c *gin.Context) {
	reportidstr := c.DefaultQuery("rid", "-1")
	uidstr := c.DefaultQuery("uid", "-1")
	typestr := c.DefaultQuery("type", "错误")
	rtypestr := c.DefaultQuery("rtype", "错误")
	idstr := c.DefaultQuery("id", "-1")
	daystr := c.DefaultQuery("day", "-2")
	ruidstr := c.DefaultQuery("ruid", "-1")
	reportid, err_rid := strconv.Atoi(reportidstr)
	uid, err_uid := strconv.Atoi(uidstr)
	id, err_id := strconv.Atoi(idstr)
	day, err_day := strconv.Atoi(daystr)
	ruid, err_ruid := strconv.Atoi(ruidstr)
	if err_rid != nil || err_uid != nil || err_id != nil || err_day != nil || err_ruid != nil || reportid == -1 || uid == -1 || id == -1 || day == -2 || ruid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if typestr == "错误" || rtypestr == "错误" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if day == -1 {
		day = 1000
	}

	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	var dataid int
	if rtypestr == "log" {
		query := "UPDATE postreports SET is_handled=1 WHERE report_id = ?"
		_, err = db.Exec(query, reportid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子封禁失败"})
			return
		}
		//查询帖子id
		query = "SELECT post_id FROM postreports WHERE report_id = ?"
		row := db.QueryRow(query, reportid)
		err = row.Scan(&dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子id失败"})
			return
		}

	} else if rtypestr == "comment" || rtypestr == "reply" {
		query := "UPDATE commentreports SET is_handled=1 WHERE report_id = ?"
		_, err = db.Exec(query, reportid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论封禁失败"})
			return
		}
		//查询评论id
		query = "SELECT comment_id FROM commentreports WHERE report_id = ?"
		row := db.QueryRow(query, reportid)
		err = row.Scan(&dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询评论id失败"})
			return
		}
	}
	//生成系统消息
	is_ok, info := MakeSysinfo(typestr, rtypestr, id, day)
	if !is_ok {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": info})
		return
	}
	query := "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	var type_r string
	if typestr == "封禁" {
		type_r = "封禁通知"
	} else if typestr == "禁言" {
		type_r = "禁言通知"
	}
	_, err = db.Exec(query, uid, type_r, info)
	//生成用户反馈

	is_ok, info = UserFeedback(typestr, day, uid)
	if !is_ok {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": info})
		return
	}
	query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	_, err = db.Exec(query, ruid, "用户反馈", info)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	//插入封禁表
	query = "INSERT INTO usermutes (user_id, type, start_time,end_time) VALUES (?, ?, ?, ?)"
	currentTime := time.Now()
	chinaTime := currentTime.Add(8 * time.Hour)
	start := chinaTime
	end := start.Add(time.Duration(day) * 24 * time.Hour)
	var startstr string
	var endstr string
	startstr = start.Format("2006-01-02 15:04:05")
	endstr = end.Format("2006-01-02 15:04:05")
	real_type := 0
	if typestr == "禁言" {
		real_type = 1
	}
	_, err = db.Exec(query, uid, real_type, startstr, endstr)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "封禁表插入失败"})
		return
	}
	if rtypestr == "log" {
		query = "DELETE FROM posts WHERE post_id = ?"
		_, err = db.Exec(query, dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子删除失败"})
			return
		}
	} else if rtypestr == "comment" || rtypestr == "reply" {
		query = "DELETE FROM comments WHERE comment_id = ?"
		_, err = db.Exec(query, dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论删除失败"})
			return
		}
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})

}

// 管理员对举报不做任何处理
func AdmIgnore(c *gin.Context) {
	reportidstr := c.DefaultQuery("rid", "-1")
	typestr := c.DefaultQuery("type", "错误")
	reportid, err_rid := strconv.Atoi(reportidstr)
	if err_rid != nil || reportid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if typestr == "错误" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	if typestr == "log" {
		query := "UPDATE postreports SET is_handled=1 WHERE report_id = ?"
		_, err = db.Exec(query, reportid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子封禁失败"})
			return
		}
	} else if typestr == "comment" {
		query := "UPDATE commentreports SET is_handled=1 WHERE report_id = ?"
		_, err = db.Exec(query, reportid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论封禁失败"})
			return
		}
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})
}

// 管理员警告
func AdmWarn(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	content := c.DefaultQuery("content", "")
	ruidstr := c.DefaultQuery("ruid", "-1")
	reportidstr := c.DefaultQuery("rid", "-1")
	typestr := c.DefaultQuery("type", "错误")
	uid, err_uid := strconv.Atoi(uidstr)
	ruid, err_ruid := strconv.Atoi(ruidstr)
	reportid, err_rid := strconv.Atoi(reportidstr)
	if err_uid != nil || err_ruid != nil || err_rid != nil || uid == -1 || ruid == -1 || reportid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if typestr == "错误" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	//删除帖子/评论
	var dataid int
	//更新举报表
	if typestr == "log" {

		query := "UPDATE postreports SET is_handled=1 WHERE report_id = ?"
		_, err = db.Exec(query, reportid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子封禁失败"})
			return
		}
		query_select := "SELECT post_id FROM postreports WHERE report_id = ?"
		row := db.QueryRow(query_select, reportid)
		err = row.Scan(&dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子id获取失败"})
			return
		}
	} else if typestr == "comment" {
		query := "UPDATE commentreports SET is_handled=1 WHERE report_id = ?"
		_, err = db.Exec(query, reportid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论封禁失败"})
			return
		}
		query_select := "SELECT comment_id FROM commentreports WHERE report_id = ?"
		row := db.QueryRow(query_select, reportid)
		err = row.Scan(&dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论id获取失败"})
			return
		}
	} else {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "类型错误"})
		return
	}
	//生成系统消息
	query := "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	_, err = db.Exec(query, uid, "警告通知", content)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	//生成用户反馈
	is_ok, info := UserFeedback("警告", 0, uid)
	if !is_ok {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": info})
		return
	}
	query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	_, err = db.Exec(query, ruid, "用户反馈", info)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	//删除帖子/评论
	if typestr == "log" {
		query = "DELETE FROM posts WHERE post_id = ?"
		_, err = db.Exec(query, dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "帖子删除失败"})
			return
		}
	} else if typestr == "comment" {
		query = "DELETE FROM comments WHERE comment_id = ?"
		_, err = db.Exec(query, dataid)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "评论删除失败"})
			return
		}
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})
}

func AdmOnlyWarn(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	content := c.DefaultQuery("content", "err")
	uid, err_uid := strconv.Atoi(uidstr)
	if err_uid != nil || uid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if content == "err" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}

	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	query := "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
	_, err = db.Exec(query, uid, "警告通知", content)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
		return
	}
	//提交事务
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})

}

func AdmOnlyBan(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	typestr := c.DefaultQuery("type", "错误")
	daystr := c.DefaultQuery("day", "-2")
	uid, err_uid := strconv.Atoi(uidstr)
	day, err_day := strconv.Atoi(daystr)
	if err_uid != nil || err_day != nil || uid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if typestr == "错误" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	if day == -1 {
		day = 1000
	}
	db_link, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	if typestr == "ban" {
		query := "INSERT INTO usermutes (user_id, type, start_time,end_time) VALUES (?, ?, ?, ?)"
		currentTime := time.Now()
		chinaTime := currentTime.Add(8 * time.Hour)
		start := chinaTime
		end := start.Add(time.Duration(day) * 24 * time.Hour)
		var startstr string
		var endstr string
		startstr = start.Format("2006-01-02 15:04:05")
		endstr = end.Format("2006-01-02 15:04:05")
		real_type := 0

		_, err = db.Exec(query, uid, real_type, startstr, endstr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "封禁表插入失败"})
			return
		}
		//插入系统消息
		info := "您的账户已被暂时封禁(" + startstr + "-" + endstr + ")。请您在未来遵守社区行为准则。"
		query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
		_, err = db.Exec(query, uid, "封禁通知", info)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
			return
		}
	} else if typestr == "restrick" {
		query := "INSERT INTO usermutes (user_id, type, start_time,end_time) VALUES (?, ?, ?, ?)"
		currentTime := time.Now()
		chinaTime := currentTime.Add(8 * time.Hour)
		start := chinaTime
		end := start.Add(time.Duration(day) * 24 * time.Hour)
		var startstr string
		var endstr string
		startstr = start.Format("2006-01-02 15:04:05")
		endstr = end.Format("2006-01-02 15:04:05")
		real_type := 1
		_, err = db.Exec(query, uid, real_type, startstr, endstr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "封禁表插入失败"})
			return
		}
		//插入系统消息
		info := "您的账户已被暂时禁言(" + startstr + "-" + endstr + ")。请您在未来遵守社区行为准则。"
		query = "INSERT INTO sysinfo (uid, type, content) VALUES (?, ?, ?)"
		_, err = db.Exec(query, uid, "禁言通知", info)
		if err != nil {
			db.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "系统消息存储失败"})
			return
		}

	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})

}
