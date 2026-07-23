/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
)

func DownloadFile(c *fiber.Ctx) error {
	var file rest.File
	id := c.Params("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)
	filePath := generateDownloadFilePath(idInt)

	return rest.
		Download(FiberContext{Ctx: c}).
		Source("./data/files" + filePath).
		Bind(&file).
		Respond()
}

func UploadFile(c *fiber.Ctx) error {
	var file rest.File

	return rest.
		Upload(FiberContext{Ctx: c}).
		Destination("./data/files").
		Bind(&file).
		Respond()
}
