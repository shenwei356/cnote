package main

type TagStat struct {
	Tag    string
	Amount int
}

type SortTagsByAmount []TagStat

func (tags SortTagsByAmount) Len() int {
	return len(tags)
}

func (tags SortTagsByAmount) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}

func (tags SortTagsByAmount) Less(i, j int) bool {
	return tags[i].Amount > tags[j].Amount
}

type SortItemsById []Item

func (items SortItemsById) Len() int {
	return len(items)
}
func (items SortItemsById) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items SortItemsById) Less(i, j int) bool {
	return items[i].ItemID < items[j].ItemID
}
