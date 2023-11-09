package image

import (
	jpeg "github.com/dsoprea/go-jpeg-image-structure"
)

func GetGeoDataFromBytes(b []byte) (data string, err error) {
	data, err = TryJpegWithBytes(b)

	if err == nil {
		return data, nil
	}
	return data, err
}

func TryJpegWithBytes(b []byte) (string, error) {
	jmp := jpeg.NewJpegMediaParser()

	intfc, err := jmp.ParseBytes(b)

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
