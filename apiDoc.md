The functions that front-end developers could use are:
(p *AddCurrPage) save() bool:
    this function is used to save the incoming data to both database. This function automatically inputs to the ExcData if date exists. However, the Currency table is manually checked for duplicates as inverse entry is not desired. Ex: there is no 'usd'->'idr' and 'idr'->'usd' in the currency table. However, the ExcData table can have this kind of entry.

func sevenDay(fromIn string, toIn string, timeIn time.Time, db *sql.DB) (float32,float32,*map[string]float32)
    this function is used to find(if exist) the 7 day range rate of the date given in the argument. it only returns 7 date data if the date exists in the database; otherwise, it only returns as many data as those in the 7-day range. it then calculates the average and variance as well.

func sevenMR(fromIn string, toIn string, db *sql.DB) (float32,float32,*map[string]float32)
    This function finds the most recent 7 data entry and return it. If not enough data point is available, it only returns as many points that are in the database for that particular currency. it then also returns the average and variance of those 7 data points.