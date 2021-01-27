package tools

import (
	"encoding/gob"
	"math"
	"os"
)

type Neural struct {
	Weights []BitMap
	sizeX   int
	sizeY   int
	bias    int
}

func (n *Neural)Init(sizeX, sizeY, bias int) *Neural {
	n.sizeX = sizeX
	n.sizeY = sizeY
	n.bias = bias
	return n
}

func (n *Neural) SetBias(bias int) {
	n.bias = bias
}

func (n *Neural) SaveWeights(fileName string){
	//Сохранение событий в файл (для удобной отладки)
	file, _ := os.Create(fileName)
	encoder := gob.NewEncoder(file)
	_=encoder.Encode(&n)
	_=file.Close()
}

func (n *Neural) LoadFromFile(fileName string) bool{
	//Сохранение событий в файл (для удобной отладки)
	file, err := os.Open(fileName)
	if err != nil {return false}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&n)
	if err != nil {return false}
	_=file.Close()
	return true
}

func (n *Neural) Test(view BitMap) int {
	for ind, wei := range n.Weights {
		if step(view,wei) > n.bias {
			return ind
		}
	}
	return -1
}

func (n *Neural) Identification(view BitMap) int {
	max := math.MinInt64
	num := -1
	for ind, wei := range n.Weights {
		w := step(view,wei)
		if w > max {
			max = w
			num = ind
		}
	}
	return num
}

func (n *Neural) Training(collections [][]BitMap, forceCountEra int) (int, int){
	qtItems:= len(collections)
	n.Weights = make([]BitMap,qtItems)

	for i := 0; i < qtItems; i ++ {
		n.Weights[i] = CreateBitMap(n.sizeX,n.sizeY)
	}

	countEra := 0
	countErrors := 0
	errors := true
	// Итерация по эпохам
	for (errors && forceCountEra <= 0) || (countEra < forceCountEra){
		countEra++
		//println(countEra)
		errors = false

		// Итерации по каждому числу
		for itemInd := range collections {
			weights := n.Weights[itemInd]

			// Итерации по представлению каждого числа
			for pilotInd, pilotViews := range collections {
				for _, view := range pilotViews {

					// Тест данного представления
					sum := step(view,weights)
					finded := false
					if sum > n.bias {
						finded = true
					}

					// Воздействие на веса, если требуется
					if itemInd == pilotInd && !finded {
						//fmt.Println("Ненаход  "+info)
						weights = UpdateWeights(view,weights,1,1)
						countErrors++
						errors = true
					}
					if itemInd != pilotInd && finded {
						weights = UpdateWeights(view,weights,1,-1)
						countErrors++
						errors = true
					}

				}
			}
		}
	}
	return countEra, countErrors
}

func step(bitMap, weightsMap BitMap) int {
	sum := 0
	for x := range bitMap {
		for y := range bitMap[x] {
			sum += int(bitMap[x][y]*weightsMap[x][y])
		}
	}
	return sum
}

func UpdateWeights(bitMap, weightsMap BitMap, triggerValue, increase int16) BitMap {
	for x := range bitMap {
		for y := range bitMap[x] {
			if bitMap[x][y] == triggerValue {
				weightsMap[x][y] = weightsMap[x][y] + increase
			}
		}
	}
	return weightsMap
}
