package ddb

type RowRequest struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

type RowResponse struct {
	Columns []TableColumn `json:"columns"`
	Values  []string      `json:"values"`
}

func (cn *Cnct) GetRow(request *RowRequest) (*RowResponse, error) {
	var (
		response RowResponse
		err      error
	)
	columns, rs, err := cn.getRow(request.Table, request.ID, true)
	if err == nil {
		response.Values = *rs
		response.Columns = *columns
		return &response, err
	}
	return nil, err
}
