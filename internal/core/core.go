package core


func remapTXfromEntityToDTO(entityTXmodel) dtomodel, err {

	dtomodel.Datetime = time.FromUnix(entityTXmodel.LogicalTime, 0).Format("2006-01-02 15:04:05")
	// .validate here as well
}