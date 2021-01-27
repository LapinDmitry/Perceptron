package tools

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

const (
	K_R = 0.349
	K_G = 0.387
	K_B = 0.264
)

type GreyMap [][]byte
type BitMap [][]int16

//**********************************************************************************************************************
// Подгон и обрезка изображения под заданное
//**********************************************************************************************************************
func CreateBitMap(xSize, ySize int) BitMap {
	bitMap := make([][]int16,xSize)
	for x := 0; x < xSize; x++ {
		bitMap[x] = make([]int16,ySize)
	}

	return bitMap
}

//**********************************************************************************************************************
// Загрузка изображения из джипега в оттенки серого
//**********************************************************************************************************************
func LoadGreyPicture(fileName string) GreyMap{
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}

	xLen := img.Bounds().Max.X
	yLen := img.Bounds().Max.Y

	// Поиск максимального
	var max float64
	for x := 0; x < xLen; x++{
		for y := 0; y < yLen; y++ {
			r,g,b,_ := img.At(x, y).RGBA()
			num := rgbToBigGrey(r,g,b)
			if num > max {
				max = num
			}
		}
	}

	k  := float64(255)/max

	pictureGray := make([][]byte,xLen)
	for x := 0; x < xLen; x++{
		pictureGray[x] = make([]byte,yLen)
		for y := 0; y < yLen; y++ {

			r,g,b,_ := img.At(x, y).RGBA()
			num := rgbToBigGrey(r,g,b) * k
			if num > 255 {
				pictureGray[x][y] = 255
			} else {
				pictureGray[x][y] = byte(num)
			}
		}
	}

	return pictureGray
}

// Преобразование RGB в оттенок серого
func rgbToBigGrey(r, g, b uint32) float64 {
	return K_R*float64(r) + K_G*float64(g) + K_B*float64(b)
}

//**********************************************************************************************************************
// Сохраняем оттенки серого как jpeg
//**********************************************************************************************************************
func SaveGreyPicture(greyMap GreyMap,fileName string, quality int) bool{
	xLen := len(greyMap)
	yLen := len(greyMap[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{xLen, yLen}

	pict := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < xLen; x++{
		for y := 0; y < yLen; y++ {
			r,g,b := greyToRGB(greyMap[x][y])
			pix := color.RGBA{r, g, b, 0xff}
			pict.Set(x,y,pix)
		}
	}

	// Somewhere in the same package
	f2, err2 := os.Create(fileName)
	if err2 != nil {
		return false
	}
	defer f2.Close()

	opt := jpeg.Options{
		Quality: quality,
	}
	err := jpeg.Encode(f2, pict, &opt)
	if err != nil {
		return false
	}
	return true
}

// Преобразование оттенока серого в ргб
func greyToRGB(pix byte) (uint8,uint8, uint8) {
	r := K_R*float32(pix)*3
	if r > 255 {
		r = 255
	}
	g := K_G*float32(pix)*3
	if g > 255 {
		g = 255
	}
	b := K_G*float32(pix)*3
	if b > 255 {
		b = 255
	}
	return	uint8(r),
		uint8(g),
		uint8(b)
}

//**********************************************************************************************************************
// Перевод из оттенков серого в карту битов (изображение пиксель может быть либо 0, либо 1)
//**********************************************************************************************************************
func GreyToBitmap(grey GreyMap,grey_k float64)BitMap{

	xLen := len(grey)
	yLen := len(grey[0])

	threshold := 0

	for x := 0; x < xLen; x++{
		for y := 0; y < yLen; y++ {
			threshold += int(grey[x][y])
		}
	}
	threshold = int(float64(threshold)*grey_k/float64((xLen*yLen)))

		intMap := make([][]int16,xLen)
	for x := 0; x < xLen; x++{

		intMap[x] = make([]int16,yLen)
		for y := 0; y < yLen; y++ {
			if grey[x][y] > byte(threshold) {
				intMap[x][y] = 0
			} else {
				intMap[x][y] = 1
			}
		}
	}
	return intMap
}

//**********************************************************************************************************************
// Перевод карты битов в оттенки серого
//**********************************************************************************************************************
func BitmapToGray(bitMap BitMap)GreyMap{
	xLen := len(bitMap)
	yLen := len(bitMap[0])

	// Поиск максимального
	var max int16= -32000
	var min int16 = 0
	for x := 0; x < xLen; x++{
		for y := 0; y < yLen; y++ {
			if bitMap[x][y] > max {
				max = bitMap[x][y]
			}
			if bitMap[x][y] < min {
				min = bitMap[x][y]
			}
		}
	}

	k  := float64(255)/float64(max-min)

	grey := make([][]byte,xLen)
	for x := 0; x < xLen; x++{
		grey[x] = make([]byte,yLen)
		for y := 0; y < yLen; y++ {
			pix := byte(float64(bitMap[x][y]-min)*k)
			if pix > 255 {
				pix = 255
			}
			grey[x][y] = 255 - pix
		}
	}
	return grey
}

//**********************************************************************************************************************
// Поиск границ самого большого объекта на фото
//**********************************************************************************************************************
func FindObjects(bitMap BitMap) []ObjectInfo {
	xLen := len(bitMap)
	yLen := len(bitMap[0])

	objects := make([]ObjectInfo,0,100)

	checkMap := make([][]bool,xLen)
	for x := 0; x < xLen; x++{
		checkMap[x] = make([]bool,yLen)
	}

	for x := range bitMap {
		for y := range bitMap[x] {
			if bitMap[x][y] != 1 || checkMap[x][y] {
				continue
			}

			info := &ObjectInfo{
				X0:       x,
				X1:       x,
				Y0:       y,
				Y1:       y,
				Count:    1,
			}

			findObject(&bitMap,&checkMap,info,x,y)

			objects = append(objects, *info)
		}
	}

	return objects
}

// Структура для удобного поиска в рекурсии
type ObjectInfo struct {
	X0       int
	X1       int
	Y0       int
	Y1       int
	Count    int
}

// Функция рекурсии, которая анализирует 4 клетки воокруг
func findObject(pBitMap*BitMap,pBCheckMap *[][]bool,info *ObjectInfo,xPix,yPix int) {
	bitMap := *pBitMap
	checkMap := *pBCheckMap

	xLen := len(bitMap)
	yLen := len(bitMap[0])

	(checkMap)[xPix][yPix] = true
	for x := xPix-1; x <= xPix+1; x++{
		for y := yPix-1; y <= yPix+1; y++ {
			if !(x == xPix && y == yPix) && (x >= 0) && (x < xLen) && (y >= 0) && (y < yLen) &&
				!(checkMap)[x][y]  && bitMap[x][y] == 1 {
				info.Count++
				if x < info.X0 {
					info.X0 = x
				}
				if x > info.X1 {
					info.X1 = x
				}
				if y < info.Y0 {
					info.Y0 = y
				}
				if y > info.Y1 {
					info.Y1 = y
				}

				findObject(pBitMap,pBCheckMap,info,x,y)
			}
		}
	}
}

func SortByDensity(pObjects *[]ObjectInfo, increase bool)  {
	objects := *pObjects
	stop := false
	for stop == false {
		stop = true

		for i := 1; i < len(objects); i++ {
			size1 := (objects[i-1].X1-objects[i-1].X0+1)*(objects[i-1].Y1-objects[i-1].Y0+1)
			size2 := (objects[i].X1-objects[i].X0+1)*(objects[i].Y1-objects[i].Y0+1)

			param1 := float64(objects[i-1].Count*objects[i-1].Count)/float64(size1)
			param2 := float64(objects[i].Count*objects[i].Count)/float64(size2)

			if (param1 > param2 && increase) ||
				(param1 < param2 && !increase) {
				stop = false
				objects[i-1], objects[i] = objects[i], objects[i-1]
			}
		}
	}
}

func SortByX(pObjects *[]ObjectInfo, increase bool)  {
	objects := *pObjects
	stop := false
	for stop == false {
		stop = true

		for i := 1; i < len(objects); i++ {
			param1 := objects[i-1].X0
			param2 := objects[i].X0

			if (param1 > param2 && increase) ||
				(param1 < param2 && !increase) {
				stop = false
				objects[i-1], objects[i] = objects[i], objects[i-1]
			}
		}
	}
}

func SortByMass(pObjects *[]ObjectInfo, increase bool)  {
	objects := *pObjects
	stop := false
	for stop == false {
		stop = true

		for i := 1; i < len(objects); i++ {
			param1 := objects[i-1].Count
			param2 := objects[i].Count

			if (param1 > param2 && increase) ||
				(param1 < param2 && !increase) {
				stop = false
				objects[i-1], objects[i] = objects[i], objects[i-1]
			}
		}
	}
}

//**********************************************************************************************************************
// Подгон и обрезка изображения под заданное
//**********************************************************************************************************************
func ResizeToStandart(bitMap BitMap,xTemplate, yTemplate int) BitMap {
	xLen := len(bitMap)
	yLen := len(bitMap[0])

	kx := float64(xTemplate)/float64(xLen)
	ky := float64(yTemplate)/float64(yLen)

	var k float64

	if kx < ky {
		k = kx
	} else {
		k = ky
	}

	bitMap = resize(bitMap,k)
	xLen = len(bitMap)
	yLen = len(bitMap[0])

	field := make([][]int16,xTemplate)
	for x := 0; x < xTemplate; x++{
		field[x] = make([]int16,yTemplate)
	}

	xPos := (xTemplate - xLen)/2
	yPos := (yTemplate - yLen)/2

	for x := 0; x < xLen; x++{
		for y := 0; y < yLen; y++{
			field[xPos+x][yPos+y] = bitMap[x][y]
		}
	}

	return field
}

// Изменение размера с коэффициентом
func resize(bitMap BitMap,k float64) BitMap {
	xLen := len(bitMap)
	yLen := len(bitMap[0])

	xNewLen := int(float64(xLen)*k)
	yNewLen := int(float64(yLen)*k)

	newBitMap := make([][]int16,xNewLen)

	for x := 0; x < xNewLen; x++ {
		xGet := int(float64(x)/k)
		if xGet < 0 {
			xGet = 0
		}
		if xGet > xLen-1 {
			xGet = xLen-1
		}

		newBitMap[x] = make([]int16,yNewLen)

		for y := 0; y < yNewLen; y++ {
			yGet := int(float64(y)/k)
			if yGet < 0 {
				yGet = 0
			}
			if yGet > yLen-1 {
				yGet = yLen-1
			}

			newBitMap[x][y] = bitMap[xGet][yGet]
		}
	}
	return newBitMap
}

//**********************************************************************************************************************
// Подгон с реформатом
//**********************************************************************************************************************
func ResizeWithReformat(bitMap BitMap,xNewLen, yNewLen int) BitMap {
	xLen := len(bitMap)
	yLen := len(bitMap[0])

	newBitMap := make([][]int16,xNewLen)
	kx := float64(xLen)/float64(xNewLen)
	ky := float64(yLen)/float64(yNewLen)

	for x := 0; x < xNewLen; x++ {
		xGet := int(float64(x)*kx)
		if xGet < 0 {
			xGet = 0
		}
		if xGet > xLen-1 {
			xGet = xLen-1
		}

		newBitMap[x] = make([]int16,yNewLen)

		for y := 0; y < yNewLen; y++ {
			yGet := int(float64(y)*ky)
			if yGet < 0 {
				yGet = 0
			}
			if yGet > yLen-1 {
				yGet = yLen-1
			}

			newBitMap[x][y] = bitMap[xGet][yGet]
		}
	}
	return newBitMap
}

//**********************************************************************************************************************
// Вырезание куска карты
//**********************************************************************************************************************
func Cut(bitMap BitMap,x0,x1,y0,y1 int) BitMap {
	xNewLen := x1-x0+1
	yNewLen := y1-y0+1

	newMap := make([][]int16,xNewLen)
	for x := 0; x < xNewLen; x++{

		newMap[x] = make([]int16,yNewLen)
		for y := 0; y < yNewLen; y++ {
			newMap[x][y] = bitMap[x0+x][y0+y]
		}
	}

	return newMap
}

//**********************************************************************************************************************
// Исправление проекции
// Подгонка точки в своей угол
//**********************************************************************************************************************
func Corner(bitMap BitMap,xCorner,yCorner, cornerNum int) BitMap {
	xLen := len(bitMap)
	yLen := len(bitMap[0])

	//if cornerNum == 1 {
	//	SaveGreyPicture(BitmapToGray(bitMap),"!.jpg",90)
	//}

	var newBitMap [][]int16
	// Первая прогонка - исправляем Х
	if (xCorner == 0 || xCorner == xLen-1) {
		newBitMap = bitMap
	} else {
		newBitMap = make([][]int16,xLen)
		for x := 0; x < xLen; x++ {
			newBitMap[x] = make([]int16,yLen)
			for y := 0; y < yLen; y++ {
				//if x == xLen-1 && y == yLen-1 &&

				xGet:= getXfromOld(xLen-1,yLen-1,x,y,xCorner,yCorner,cornerNum)
				newBitMap[x][y] = bitMap[xGet][y]
			}
		}

		// Корректируем координаты угла
		if cornerNum == 2 || cornerNum == 3 {
			xCorner = xLen-1
		} else {
			xCorner = 0
		}
	}

	//SaveGreyPicture(BitmapToGray(newBitMap),strconv.Itoa(cornerNum)+" X.jpg",90)

	// Вторая прогонка - исправляем Y
	if (yCorner == 0 || yCorner == yLen-1) {
		return newBitMap
	}

	newBitMap2 := make([][]int16,xLen)
	for x := 0; x < xLen; x++ {
		newBitMap2[x] = make([]int16,yLen)
		for y := 0; y < yLen; y++ {
			yGet:= getYfromOld(xLen-1,yLen-1,x,y,xCorner,yCorner,cornerNum)

			newBitMap2[x][y] = newBitMap[x][yGet]
		}
	}

	//SaveGreyPicture(BitmapToGray(newBitMap2),strconv.Itoa(cornerNum)+" Y.jpg",90)

	return newBitMap2
}

// Высчитываение Х
func getXfromOld(xMax, yMax,x,y,xCorner,yCorner, cornerNum int) int{
	var kx,bx float64
	var newX int
	switch cornerNum {
	case 1:
		kx, bx = getLineFunc(0, yMax, xCorner, yCorner)
	case 2:
		kx, bx = getLineFunc(xMax, yMax, xCorner, yCorner)
	case 3:
		kx, bx = getLineFunc(xMax, 0, xCorner, yCorner)
	case 4:
		kx, bx = getLineFunc(0, 0, xCorner, yCorner)
	}
	if cornerNum == 2 || cornerNum == 3 {
		// Искажения по X
		xBorder := (float64(y)-bx)/kx
		kxDistortion := xBorder/float64(xMax)
		newX = int(float64(x)*kxDistortion)
	} else {
		xBorder := float64(xMax) - (float64(y)-bx)/kx
		kxDistortion := xBorder/float64(xMax)
		newX = xMax - int(float64(xMax-x)*kxDistortion)
	}

	if newX < 0 {
		newX = 0
	}
	if newX > xMax-1 {
		newX = xMax -1
	}

	return newX
}

// Высчитываение Y
func getYfromOld(xMax, yMax,x,y,xCorner,yCorner, cornerNum int) int{
	var ky,by float64
	var newY int
	switch cornerNum {
	case 1:
		ky, by = getLineFunc(xMax,0,xCorner,yCorner)
	case 2:
		ky, by = getLineFunc(0,0,xCorner,yCorner)
	case 3:
		ky, by = getLineFunc(0, yMax,xCorner,yCorner)
	case 4:
		ky, by = getLineFunc(xMax, yMax,xCorner,yCorner)
	}

	if cornerNum == 3 || cornerNum == 4 {
		// Искажения по Y
		yBorder := float64(x)*ky+by
		kyDistortion := yBorder/float64(yMax)
		newY = int(float64(y)* kyDistortion)
	} else {
		yBorder := float64(yMax) - (float64(x)*ky+by)
		kyDistortion := yBorder/float64(yMax)
		newY = yMax -int(float64(yMax-y)* kyDistortion)
	}

	if newY < 0 {
		newY = 0
	}
	if newY > yMax-1 {
		newY = yMax -1
	}

	return newY
}

// Узнаём параметры линейной функции
func getLineFunc(x0,y0,x1,y1 int) (float64, float64) {
	k := float64(y1-y0)/float64(x1-x0)
	b := float64(y0) - k*float64(x0)
	return k, b
}