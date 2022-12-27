package ddb

type DeleteRowRequest struct {
	Table string `json:"table" form:"table"`
	ID    string `json:"id" form:"id"`
}

type DeleteRowResponse struct {
}

func (cn *Cnct) DeleteRow(request *DeleteRowRequest) (*DeleteRowResponse, error) {
	var (
		response DeleteRowResponse
		err      error
		id       string
	)
	if len(request.ID) > 0 {
		id = request.ID
	}
	_, err = cn.Db.Exec("DELETE FROM "+request.Table+" WHERE id = ? ", id)
	if err == nil {
		return &response, err
	}
	return nil, err
}
