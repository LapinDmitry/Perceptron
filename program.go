package main

import (
	"fmt"
	_ "image/jpeg"
	"math"
	"os"
	"strconv"
	"tools"
) //register JPEG decoder

var SIZE_FIELD_X int
var SIZE_FIELD_Y int

var	SIZE_CHAR_X int
var	SIZE_CHAR_Y int

var ROOT_TRAINING string = "io/тренировочные/"
var ROOT_IN string	= "io/ввод/"
var ROOT_STEPS string = "io/нарезка/"
var ROOT_WEIGHTS string = "io/веса/"
var FILE_NAME string = "io/нейро.сеть"

var GREY_FACTOR float64
var TRAINING_COUNT int
var COUNT_ERA int
var BIAS int

var NUM_LEN = 10

var JPG = ".jpg"


func readConf(fileName string) bool {
	println(fmt.Sprintf("Загрузка файла конфигурации %s...",fileName))
	file, err := os.Open(fileName)
	if err != nil {
		println("Ошибка! Ошибка чтения конфигурации.")
		return false
	}
	errs := make([]error,8)
	var spaceLine string
	_,_=fmt.Fscanln(file,&spaceLine);
	_,errs[0]=fmt.Fscanln(file,&SIZE_FIELD_X)
	_,errs[1]=fmt.Fscanln(file,&SIZE_FIELD_Y)

	_,_=fmt.Fscanln(file,&spaceLine);
	_,errs[2]=fmt.Fscanln(file,&SIZE_CHAR_X)
	_,errs[3]=fmt.Fscanln(file,&SIZE_CHAR_Y)

	_,_=fmt.Fscanln(file,&spaceLine);
	_,errs[4]=fmt.Fscanln(file,&GREY_FACTOR)

	_,_=fmt.Fscanln(file,&spaceLine);
	_,errs[5]=fmt.Fscanln(file,&TRAINING_COUNT)
	_,_=fmt.Fscanln(file,&spaceLine);
	_,errs[6]=fmt.Fscanln(file,&COUNT_ERA)
	_,_=fmt.Fscanln(file,&spaceLine);
	_,errs[7]=fmt.Fscanln(file,&BIAS)



	_=file.Close()
	for _,e := range errs {
		if e != nil {
			println("Ошибка! Ошибка чтения конфигурации.")
			return false
		}
	}
	println("Конфигурация успешно прочитана.")
	return true
}

func training(neural *tools.Neural,forceCountEra, bias int) {
	neural.SetBias(bias)

	println(fmt.Sprintf("Количество эр:%d \nПорог срабатывания:%d \nОбучение...",forceCountEra,bias))

	collections := make([][]tools.BitMap,10)
	for numInd := range collections {
		collections[numInd] = make([]tools.BitMap,0)
	}

	count := 0
	for i := 0; i < TRAINING_COUNT; i++ {
		views := getNumber(ROOT_TRAINING,strconv.Itoa(i))

		for numInd := range collections {
			collections[numInd] = append(collections[numInd],views[numInd])
		}
	}

	countEra, countErr := neural.Training(collections,forceCountEra)
	println()
	println("=====================================================================================")
	println(fmt.Sprintf("Нейросеть обучена количество эр: %d, количество корректировок %d.",countEra,countErr))
	println("=====================================================================================")
	for _,w := range neural.Weights {
		count++
		tools.SaveGreyPicture(tools.BitmapToGray(w),ROOT_WEIGHTS+strconv.Itoa(count)+JPG,90)
	}

}

func getDistanse(x0,y0,x1,y1 int) int {
	xx := float64(x0-x1)
	yy := float64(y0-y1)
	return int(math.Sqrt(xx*xx+yy*yy))
}

func min(num1, num2 int) int {
	if num1 < num2 {
		return num1
	} else {
		return num2
	}
}

func max(num1, num2 int) int {
	if num1 > num2 {
		return num1
	} else {
		return num2
	}
}

func findFourCorners(bitMap tools.BitMap) (tools.ObjectInfo,tools.ObjectInfo,tools.ObjectInfo,tools.ObjectInfo) {
	xLen := len(bitMap)
	yLen := len(bitMap[0])

		// Поиск 4 углов
		objects := tools.FindObjects(bitMap)
		tools.SortByDensity(&objects, false)

		leftUp := objects[0]
		leftDown := objects[0]
		righUp := objects[0]
		righDown := objects[0]
		for _, ob := range objects[:4] {
			if ob.X0+ob.Y0 < leftUp.X0+leftUp.Y0 {
				leftUp = ob
			}
			if ob.X0+(yLen-ob.Y1) < leftDown.X0+(yLen-leftDown.Y1) {
				leftDown = ob
			}
			if (xLen-ob.X1)+ob.Y0 < (xLen-righUp.X1)+righUp.Y0 {
				righUp = ob
			}
			if ob.X1+ob.Y1 > righDown.X1+righDown.Y1 {
				righDown = ob
			}
		}

	return leftUp, righUp, righDown, leftDown
}

func restoringProjection(bitMap tools.BitMap) tools.BitMap{
	leftUp, righUp, righDown, leftDown := findFourCorners(bitMap)
	x0 := min(leftDown.X0,leftUp.X0)
	x1 := max(righDown.X1,righUp.X1)
	y0 := min(leftUp.Y0,righUp.Y0)
	y1 := max(leftDown.Y1,righDown.Y1)

	bitMap = tools.Cut(bitMap,x0,x1,y0,y1)

	//count := 10000
	//tools.SaveGreyPicture(tools.BitmapToGray(bitMap),strconv.Itoa(count)+".jpg",90)

	for cornerNum := 1; cornerNum <= 4; cornerNum++ {
		leftUp, righUp, righDown, leftDown := findFourCorners(bitMap)

		switch cornerNum {
		case 1:bitMap = tools.Corner(bitMap,leftUp.X0,leftUp.Y0,1)
		case 2:bitMap = tools.Corner(bitMap,righUp.X1,righUp.Y0,2)
		case 3:bitMap = tools.Corner(bitMap,righDown.X1,righDown.Y1,3)
		case 4:bitMap = tools.Corner(bitMap,leftDown.X0,leftDown.Y1,4)
		}

		//count++
		//tools.SaveGreyPicture(tools.BitmapToGray(bitMap),strconv.Itoa(count)+".jpg",90)
	}

	//count++
	//tools.SaveGreyPicture(tools.BitmapToGray(bitMap),strconv.Itoa(count)+".jpg",90)

	leftUp, righUp, righDown, leftDown = findFourCorners(bitMap)
	x0 = max(leftDown.X1,leftUp.X1)
	x1 = min(righDown.X0,righUp.X0)
	y0 = max(leftUp.Y1,righUp.Y1)
	y1 = min(leftDown.Y0,righDown.Y0)
	bitMap = tools.Cut(bitMap,x0,x1,y0,y1)

	//count++
	//tools.SaveGreyPicture(tools.BitmapToGray(bitMap),strconv.Itoa(count)+".jpg",90)

	return bitMap
}

func getNumber(root, fileName string) []tools.BitMap {
	pict := tools.LoadGreyPicture(root+fileName+JPG)
	bitMap := tools.GreyToBitmap(pict,GREY_FACTOR)

	out := ROOT_STEPS+fileName

	tools.SaveGreyPicture(tools.BitmapToGray(bitMap),out+"-%"+JPG,90)

	bitMap = restoringProjection(bitMap)
	bitMap = tools.ResizeWithReformat(bitMap,SIZE_FIELD_X,SIZE_FIELD_Y)

	tools.SaveGreyPicture(tools.BitmapToGray(bitMap),out+"--%"+JPG,90)

	objects := tools.FindObjects(bitMap)
	numObjects := make([]tools.ObjectInfo,0,NUM_LEN)
	for _, o := range objects {
		k := float64(o.Count)/float64(SIZE_FIELD_X*SIZE_FIELD_Y)
		if k >= 0.0005 && o.Y0 > SIZE_FIELD_Y/2 {
			numObjects = append(numObjects, o)
			//println(fmt.Sprintf("Масса: %d  dx:%d  dy:%d  k:%f",o.Count,o.X1-o.X0,o.Y1-o.Y0,k))
		}
	}

	tools.SortByX(&numObjects,true)
	number := make([]tools.BitMap,10)
	for ind, obj := range numObjects {
		number[ind] = tools.Cut(bitMap,obj.X0,obj.X1,obj.Y0,obj.Y1)
		number[ind] = tools.ResizeToStandart(number[ind],SIZE_CHAR_X,SIZE_CHAR_Y)

		tools.SaveGreyPicture(tools.BitmapToGray(number[ind]),out+"----%"+strconv.Itoa(ind)+JPG,90)
	}
	return number
}

func test(neural *tools.Neural,fileName string, numTrue string) {
	number := getNumber(ROOT_IN,fileName)

	res := ""
	for _, num := range number {
		n := neural.Identification(num)
		res += strconv.Itoa(n)
	}

	println()
	println("=====================================================================================")
	println(fileName+JPG)
	if res == numTrue {
		println(fmt.Sprintf("Распознало: %s, ВЕРНО",res))
	} else {
		println(fmt.Sprintf("Распознало: %s, НЕВЕРНО",res))
	}
	println("=====================================================================================")
}

func main() {
	ans := ""
	if ok := readConf("init.txt");ok == false {
		println("Выход!")
		_,_ = fmt.Fscan(os.Stdin, &ans)
		return
	}
	neural := tools.Neural{}

	if ok := neural.LoadFromFile(FILE_NAME);ok == true {
		println(fmt.Sprintf("Нейросеть успешно загружена из файла %s.",FILE_NAME))
	} else {
		println(fmt.Sprintf("Не удалось загрузить файла %s.",FILE_NAME))
	}

	neural.Init(SIZE_CHAR_X,SIZE_CHAR_Y,SIZE_FIELD_Y*SIZE_FIELD_X/100)


	for ans != "в"{
		println()
		fmt.Println("Команды:")
		fmt.Println("т (количество эр) (порог срабатывания)     - тренировать у у - аргументы по умолчанию (из файла конфигурации)")
		fmt.Println("и (имя файла без .jpg) (число на картинке) - испытать")
		fmt.Println("с (имя файла без .jpg)                     - сохранить")
		fmt.Println("о (имя файла без .jpg)                     - открыть")
		fmt.Println("в - выход")

		name := ""
		_,_ = fmt.Fscan(os.Stdin, &ans)
		switch ans {
		case "т":
			var count, bias int

			_,_ = fmt.Fscan(os.Stdin, &name)
			if name == "у" {
				count = COUNT_ERA
			} else {
				count, _ = strconv.Atoi(name)
			}

			_,_ = fmt.Fscanln(os.Stdin, &name)
			if name == "у" {
				bias = BIAS
			}else {
				bias, _ = strconv.Atoi(name)
			}

			training(&neural,count,bias)
			neural.SaveWeights(FILE_NAME)

		case "и":
			_,_ = fmt.Fscan(os.Stdin,&name)
			var num string
			_,_ = fmt.Fscanln(os.Stdin,&num)
			test(&neural,name,num)
		case "с":

			_,_ = fmt.Fscanln(os.Stdin,&name)
			neural.SaveWeights(name)
			println(fmt.Sprintf("Сохранено в %s",name))

		case "о":
			_,_ = fmt.Fscanln(os.Stdin,&name)
			neural.LoadFromFile(name)
			println(fmt.Sprintf("Загружено из %s",name))

		default:
			_,_ = fmt.Fprintf(os.Stdout,"Неизвестная команда")
		}

	}
}