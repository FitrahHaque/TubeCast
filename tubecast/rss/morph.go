package rss

func StationToStationMeta(station *Station) metaStation {
	items := make([]metaStationItem, len(station.Items))
	for i, item := range station.Items {
		items[i] = metaStationItem{
			stationItem: &station.Items[i],
		}
	}
}

func StationItemToStation(station *StationItem) *metaStationItem {

}
