package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 检查用户是否存在
func UserExists(db *sql.DB, userID string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE user_id = ?"
	var count int
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查用户是否存在失败：%v")
	}
	return count > 0, nil
}

// 关注博主
func PerformFollow(tx *sql.Tx, followerID, followedID string, actionType string) error {
	if actionType == "1" {
		//检查是否早已经关注
		check1 := "SELECT COUNT(*) FROM userfollows WHERE follower_id = ? AND followed_id = ?"
		var count int
		err1 := tx.QueryRow(check1, followerID, followedID).Scan(&count)
		if err1 != nil {
			return fmt.Errorf("检查是否已经关注失败：%v", err1)
		}
		if count == 1 {
			return fmt.Errorf("已经关注该博主，重复操作")
		}
		//添加关注
		query := "INSERT INTO userfollows (follower_id, followed_id,follow_time) VALUES (?, ?,?)"
		_, err := tx.Exec(query, followerID, followedID, time.Now())
		if err != nil {
			fmt.Errorf("关注失败：%v", err)
		}
	} else if actionType == "0" {
		//检查是否已经关注
		check1 := "SELECT COUNT(*) FROM userfollows WHERE follower_id = ? AND followed_id = ?"
		var count int
		err1 := tx.QueryRow(check1, followerID, followedID).Scan(&count)
		if err1 != nil {
			return fmt.Errorf("检查是否已经关注失败：%v", err1)
		}
		if count == 0 {
			return fmt.Errorf("没有关注该博主，无效操作")
		}
		//取消关注
		query := "DELETE FROM userfollows WHERE follower_id = ? AND followed_id = ?"
		_, err := tx.Exec(query, followerID, followedID)
		if err != nil {
			return fmt.Errorf("取消关注失败：%v", err)
		}
	} else {
		return fmt.Errorf("无效的操作类型")
	}
	return nil

}

func HandleFollowAction(c *gin.Context) {
	uid := c.Query("uid")
	attionuid := c.Query("attionuid")
	actionType := c.Query("type")

	if uid == "" || attionuid == "" || actionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "参数uid或attionuid或type不能为空"})
		return
	}

	if actionType != "1" && actionType != "0" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "type参数无效"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}

	defer db.Close()

	//检查用户是否存在
	exist1, err1 := UserExists(db, uid)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err1.Error()})
		return
	}

	exist2, err2 := UserExists(db, attionuid)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err2.Error()})
		return
	}
	if !exist2 {
		if exist1 {
			failReason := fmt.Sprintf("用户(%s)不存在", attionuid)
			c.JSON(http.StatusBadRequest, gin.H{
				"isok":       false,
				"failreason": failReason,
			})
			return
		} else {
			failReason := fmt.Sprintf("用户(%s)和(%s)都不存在", uid, attionuid)
			c.JSON(http.StatusBadRequest, gin.H{
				"isok":       false,
				"failreason": failReason,
			})
			return
		}
	}
	if !exist1 {
		failReason := fmt.Sprintf("用户(%s)不存在", uid)
		c.JSON(http.StatusBadRequest, gin.H{
			"isok":       false,
			"failreason": failReason,
		})
		return
	}

	//事务开启
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "开启事务失败"})
		return
	}

	err = PerformFollow(tx, uid, attionuid, actionType)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})
}

// 添加回复的点赞
func addLikeReply(tx *sql.Tx, uid, comID string, replyID string) error {
	//查看是否已经点赞
	query := "SELECT COUNT(*) FROM CommentLikes WHERE comment_id=? AND liker_id=?"
	var count int
	err := tx.QueryRow(query, replyID, uid).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询点赞失败：%v", err)
	}
	if count > 0 {
		return fmt.Errorf("已经点赞过了,无法再次点赞")
	}
	//插入点赞记录
	insertQuery := "INSERT INTO CommentLikes (comment_id, liker_id,like_time) VALUES (?, ?,?)"
	_, err = tx.Exec(insertQuery, replyID, uid, time.Now())
	if err != nil {
		return fmt.Errorf("插入点赞记录失败：%v", err)
	}

	//更新点赞数量
	updateQuery := "UPDATE Comments SET like_count=like_count+1 WHERE comment_id=? AND parent_comment_id=?"
	_, err = tx.Exec(updateQuery, replyID, comID)
	if err != nil {
		return fmt.Errorf("更新点赞数量失败：%v", err)
	}
	return nil
}

// 取消回复的点赞
func cancelLikeReply(tx *sql.Tx, uid, comID string, replyID string) error {
	//查看是否已经点赞
	query := "SELECT COUNT(*) FROM CommentLikes WHERE comment_id=? AND liker_id=?"
	var count int
	err := tx.QueryRow(query, replyID, uid).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询点赞失败：%v", err)
	}
	if count == 0 {
		return fmt.Errorf("未点赞，无法取消点赞")
	}

	//删除点赞记录
	deleteQuery := "DELETE FROM CommentLikes WHERE comment_id=? AND liker_id=?"
	_, err = tx.Exec(deleteQuery, replyID, uid)
	if err != nil {
		return fmt.Errorf("删除点赞记录失败：%v", err)
	}

	//更新点赞数量
	updateQuery := "UPDATE Comments SET like_count=like_count-1 WHERE comment_id=? AND parent_comment_id=?"
	_, err = tx.Exec(updateQuery, replyID, comID)
	if err != nil {
		return fmt.Errorf("更新点赞数量失败：%v", err)
	}
	return nil
}

// 喜欢（点赞）回复接口
func LikeReply(c *gin.Context) {
	uid := c.Query("uid")
	comID := c.Query("comid")
	actionType := c.Query("type")
	logID := c.Query("logid")
	replyID := c.Query("replyid")
	if uid == "" || comID == "" || actionType == "" || logID == "" || replyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少参数"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}
	if actionType == "1" {
		err = addLikeReply(tx, uid, comID, replyID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else if actionType == "0" {
		err = cancelLikeReply(tx, uid, comID, replyID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的操作类型"})
		return
	}

	//提交事务
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true})
}

// 举报评论/回复
func ReportComment(tx *sql.Tx, uid, comID, reason string) error {
	if comID == "" {
		return fmt.Errorf("评论或回复ID不能为空")
	}

	query := "INSERT INTO CommentReports (reporter_id,comment_id,reason,report_time) VALUES (?,?,?,?)"
	_, err := tx.Exec(query, uid, comID, reason, time.Now())
	if err != nil {
		return fmt.Errorf("举报失败:sql语句插入失败")
	}
	return nil
}

// 举报帖子
func ReportPost(tx *sql.Tx, uid, postID, reason string) error {
	if postID == "" {
		return fmt.Errorf("帖子ID不能为空")
	}
	query := "INSERT INTO PostReports (reporter_id,post_id,reason,report_time) VALUES (?,?,?,?)"
	_, err := tx.Exec(query, uid, postID, reason, time.Now())
	if err != nil {
		return fmt.Errorf("举报失败：sql插入举报帖子失败")
	}
	return nil
}

// 举报接口
func HandleReport(c *gin.Context) {
	reportType := c.Query("type")
	uid := c.Query("uid")

	if reportType == "" || uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少参数"})
		return
	}

	var req model.ReportRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求体格式错误"})
		return
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	//开启事务
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}

	if reportType == "log" {
		err = ReportPost(tx, uid, req.LogID, req.Reason)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else if reportType == "reply" {
		err = ReportComment(tx, uid, req.ReplyID, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else if reportType == "comment" {
		err = ReportComment(tx, uid, req.ComID, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "举报类型无效"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true})

}