package bridge

type (
	HTTPtypeDataSocket struct {
		Task      *HttpTask
		ISentToDB bool
		CRUDtype  int8
	}

	HttpTask struct {
		ID          uint64
		Title       string
		Description string
		CreatedAt   string
		UpdatedAt   string
	}
)

type (
	DBtypeDataSocket struct {
		Task     *DbTask
		NewTask  bool
		CRUDtype int8
	}
	DbTask struct {
	}
)

func Run(HTTPanya any, DBanya any) {
	go func(dataType any) {
		for {
			switch HTTPanya.(type) {
			case *HTTPtypeDataSocket:
				data := HTTPanya.(*HTTPtypeDataSocket)
				if data.ISentToDB {

				} else {
					continue
				}
			}
		}
	}(HTTPanya)
}

func HttpToDbMutator() {

}
