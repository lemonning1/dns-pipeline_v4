package handler

import (
	"net/http"
	"shared/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DNSHandler struct {
	service DNSGet
}
type DNSGet interface {
	GetDNSRecords(domain string, qr *int, page model.PageParams) (*model.PageResult, error)
	GetTopDomains(ddl string) ([]model.TopResult, error)
}

func NewDNSHandler(service DNSGet) *DNSHandler {
	return &DNSHandler{service: service}
}

func (h *DNSHandler) GetDNSRecords(c *gin.Context) {

	var params struct {
		Domain string `form:"domain"`
		QR     *int   `form:"qr"`
		model.PageParams
	}
	if err := c.ShouldBind(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	records, err := h.service.GetDNSRecords(params.Domain, params.QR, params.PageParams)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (h *DNSHandler) GetTopDomains(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请输入整型数字！",
		})
		return
	}

	if hours < 1 || hours > 24 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请输入1-24中的整数！",
		})
		return
	}
	items, err := h.service.GetTopDomains(hoursStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)

}
