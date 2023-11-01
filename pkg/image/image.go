package image

import (
	"fmt"

	"errors"

	"github.com/dsoprea/go-exif/v2"
	exifcommon "github.com/dsoprea/go-exif/v2/common"
	jpeg "github.com/dsoprea/go-jpeg-image-structure"
	png "github.com/dsoprea/go-png-image-structure"
	// "io/ioutil"
)

func GetGeoData(path string) (data string, err error) {

	// b, err := ioutil.ReadFile(path)

	// if err != nil {
	// return "", "", err
	// }

	data, err = TryJpeg(path)

	if err == nil {
		return data, nil
	}

	// data, err = TryPng(path)

	// if err == nil {
	// 	return data, nil
	// }

	return data, err
}

func TryJpeg(path string) (string, error) {
	jmp := jpeg.NewJpegMediaParser()

	intfc, err := jmp.ParseFile(path)

	if err != nil {
		return "", err
	}

	sl := intfc.(*jpeg.SegmentList)

	_, exifSegment, err := sl.FindExif()

	if err != nil {
		return "", err
	}

	exifs, err := exifSegment.FlatExif()

	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	return extractGPSData(exifs)
}

func extractGPSData(exifs []exif.ExifTag) (string, error) {
	tagMap := map[string]exif.ExifTag{}

	foundTagsSet := map[string]struct{}{}
	for _, exif := range exifs {
		if exif.TagName == "GPSLatitude" || exif.TagName == "GPSLongitude" || exif.TagName == "GPSLongitudeRef" || exif.TagName == "GPSLatitudeRef" {
			fmt.Printf("tagName: %v, tagValue: %T\n", exif.TagName, exif.Value)
			foundTagsSet[exif.TagName] = struct{}{}
			tagMap[exif.TagName] = exif

		}
	}

	if len(foundTagsSet) == 4 {
		latTag := tagMap["GPSLatitude"].Value.([]exifcommon.Rational)
		latTagRef := tagMap["GPSLatitudeRef"].Value.(string)

		lat := convertDMSToDD(float64(latTag[0].Numerator/latTag[0].Denominator),
			float64(latTag[1].Numerator/latTag[1].Denominator),
			float64(latTag[2].Numerator/latTag[2].Denominator),
			latTagRef)

		longTag := tagMap["GPSLongitude"].Value.([]exifcommon.Rational)
		longTagRef := tagMap["GPSLongitudeRef"].Value.(string)

		long := convertDMSToDD(float64(longTag[0].Numerator/longTag[0].Denominator),
			float64(longTag[1].Numerator/longTag[1].Denominator),
			float64(longTag[2].Numerator/longTag[2].Denominator),
			longTagRef)

		fmt.Printf("lat: %v, long: %v\n", lat, long)

		return fmt.Sprintf("Latitude: %v Longitude: %v", lat, long), nil
	} else {
		return "", errors.New("GPS EXIF data not found")
	}
}
func convertDMSToDD(degrees, minutes, seconds float64, direction string) float64 {
	var dd = degrees + minutes/60 + seconds/(60*60)

	if direction == "S" || direction == "W" {
		dd = dd * -1
	} // Don't do anything for N or E
	return dd
}

func TryPng(path string) (string, error) {
	pmp := png.NewPngMediaParser()

	intfc, err := pmp.ParseFile(path)

	if err != nil {
		return "", err
	}

	sl := intfc.(*png.ChunkSlice)

	pngExif, _, err := sl.Exif()

	if err != nil {
		return "", err
	}

	for _, tagEntry := range pngExif.Entries {
		value, _ := tagEntry.Value()

		fmt.Printf("%v %T", tagEntry.TagName(), value)
	}

	// index := sl.Index()

	// fmt.Printf("index: %#v\n", index)

	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	return "", nil
	// return extractGPSData(exifs)
}
