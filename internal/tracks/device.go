package tracks

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
)

func extractDevice(activity *filedef.Activity) string {
	if activity == nil {
		return ""
	}

	for _, info := range activity.DeviceInfos {
		if info == nil || info.DeviceIndex != typedef.DeviceIndexCreator {
			continue
		}
		if name := formatDeviceName(info.Manufacturer, info.Product, info.ProductName); name != "" {
			return name
		}
	}

	return formatDeviceName(
		activity.FileId.Manufacturer,
		activity.FileId.Product,
		activity.FileId.ProductName,
	)
}

func formatDeviceName(manufacturer typedef.Manufacturer, product uint16, productName string) string {
	productName = strings.TrimSpace(productName)
	if productName != "" {
		return combineWithManufacturer(manufacturer, productName)
	}
	return deviceLabel(manufacturer, product)
}

func combineWithManufacturer(manufacturer typedef.Manufacturer, productName string) string {
	productName = strings.TrimSpace(productName)
	if productName == "" {
		return ""
	}

	brand := brandPrefix(manufacturer)
	if brand == "" {
		return productName
	}
	if strings.Contains(strings.ToLower(productName), strings.ToLower(brand)) {
		return productName
	}
	return brand + " " + productName
}

func brandPrefix(manufacturer typedef.Manufacturer) string {
	if prefix, ok := manufacturerBrand[manufacturer]; ok {
		return prefix
	}
	label := manufacturerLabel(manufacturer)
	if label == "" {
		return ""
	}
	return strings.Fields(label)[0]
}

var manufacturerBrand = map[typedef.Manufacturer]string{
	typedef.ManufacturerWahooFitness: "Wahoo",
	typedef.ManufacturerGarmin:       "Garmin",
	typedef.ManufacturerPolarElectro: "Polar",
	typedef.ManufacturerSuunto:       "Suunto",
	typedef.ManufacturerCoros:        "Coros",
	typedef.ManufacturerBryton:       "Bryton",
	typedef.ManufacturerHammerhead:   "Hammerhead",
	typedef.ManufacturerLezyne:       "Lezyne",
	typedef.ManufacturerStagesCycling: "Stages",
	typedef.ManufacturerTacx:         "Tacx",
}

func deviceLabel(manufacturer typedef.Manufacturer, product uint16) string {
	if manufacturer == typedef.ManufacturerInvalid && product == 0 {
		return ""
	}

	info := mesgdef.DeviceInfo{
		Manufacturer: manufacturer,
		Product:      product,
	}
	_, productValue := info.GetProduct()
	productName := productLabel(productValue)
	manufacturerName := manufacturerLabel(manufacturer)

	switch {
	case manufacturerName == "" && productName == "":
		return ""
	case manufacturerName == "":
		return productName
	case productName == "":
		return manufacturerName
	case strings.Contains(strings.ToLower(productName), strings.ToLower(manufacturerName)):
		return productName
	default:
		return combineWithManufacturer(manufacturer, productName)
	}
}

func manufacturerLabel(manufacturer typedef.Manufacturer) string {
	if manufacturer == typedef.ManufacturerInvalid {
		return ""
	}
	return titleWords(strings.ReplaceAll(manufacturer.String(), "_", " "))
}

func productLabel(value any) string {
	switch product := value.(type) {
	case typedef.GarminProduct:
		if product == 0 {
			return ""
		}
		return titleWords(strings.ReplaceAll(product.String(), "_", " "))
	case typedef.FaveroProduct:
		if product == 0 {
			return ""
		}
		return titleWords(strings.ReplaceAll(product.String(), "_", " "))
	case uint16:
		return ""
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" || text == "0" {
			return ""
		}
		return titleWords(strings.ReplaceAll(text, "_", " "))
	}
}

func titleWords(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.Fields(value)
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
