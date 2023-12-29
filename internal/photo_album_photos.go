package internal

import (
	"encoding/json"
	"fmt"
)

func (r *PhotoAlbum) PhotosIter(startOffset int64) PhotosIterNext {
	if r.Direction == "DESCENDING" {
		startOffset = r.Size() - 1 - startOffset
	}
	return newPhotosIterNext(r, startOffset)
}

func (r *PhotoAlbum) GetPhotosByOffset(offset, limit int64) ([]*PhotoAsset, error) {
	var assets []*PhotoAsset

	text, err := r.service.icloud.request(&rawReq{
		Method:  "POST",
		URL:     fmt.Sprintf("%s/records/query", r.service.serviceEndpoint),
		Querys:  r.service.querys,
		Headers: r.service.icloud.getCommonHeaders(map[string]string{}),
		Body:    r.listQueryGenerate(offset, limit, r.ListType, r.Direction, r.QueryFilter),
	})
	if err != nil {
		return nil, fmt.Errorf("get album photos failed, err: %w", err)
	}
	res := new(getPhotosResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return nil, fmt.Errorf("get album photos unmarshal failed, err: %w", err)
	}

	var masterRecords []*photoRecord
	assetRecords := map[string]*photoRecord{}
	for _, record := range res.Records {
		if record.RecordType == "CPLAsset" {
			masterID := record.Fields.MasterRef.Value.RecordName
			assetRecords[masterID] = record
		} else if record.RecordType == "CPLMaster" {
			masterRecords = append(masterRecords, record)
		}
	}

	for _, masterRecord := range masterRecords {
		assets = append(assets,
			r.service.newPhotoAsset(masterRecord, assetRecords[masterRecord.RecordName]),
		)
	}

	return assets, nil
}

func (r *PhotoAlbum) GetPhotosByCount(count int) ([]*PhotoAsset, error) {
	offset := int64(0)
	if r.Direction == "DESCENDING" {
		offset = r.Size() - 1
	}

	var assets []*PhotoAsset
	for {
		tmp, err := r.GetPhotosByOffset(offset, 200)
		if err != nil {
			return nil, err
		}
		if len(tmp) == 0 {
			break
		}
		for _, v := range tmp {
			assets = append(assets, v)
			if len(assets) >= count {
				return assets, nil
			}
		}
		offset = r.calOffset(offset, int64(len(tmp)))
	}

	return assets, nil
}

func (r *PhotoAlbum) WalkPhotos(offset int64, f func(offset int64, assets []*PhotoAsset) error) error {
	size := r.Size()
	if r.Direction == "DESCENDING" {
		offset = size - 1 - offset
	}
	for {
		tmp, err := r.GetPhotosByOffset(offset, 200)
		if err != nil {
			return err
		}
		fmt.Printf("[icloudgo] [walk_photo] name: %s, offset: %d, size=%d, got=%d, desc=%v\n", r.Name, offset, size, len(tmp), r.Direction == "DESCENDING")
		if len(tmp) == 0 {
			break
		}
		offset = r.calOffset(offset, int64(len(tmp)))

		if err := f(offset, tmp); err != nil {
			return err
		}
	}
	return nil
}

func (r *PhotoAlbum) calOffset(offset, lastAssetLen int64) int64 {
	if r.Direction == "DESCENDING" {
		offset = offset - lastAssetLen
	} else {
		offset = offset + lastAssetLen
	}
	return offset
}

func (r *PhotoAlbum) listQueryGenerate(offset, limit int64, listType string, direction string, queryFilter []*folderMetaDataQueryFilter) any {
	res := map[string]any{
		"query": map[string]any{
			"filterBy": append([]*folderMetaDataQueryFilter{
				{
					FieldName:  "startRank",
					FieldValue: &folderTypeValue{Type: "INT64", Value: offset},
					Comparator: "EQUALS",
				},
				{
					FieldName:  "direction",
					FieldValue: &folderTypeValue{Type: "STRING", Value: direction},
					Comparator: "EQUALS",
				},
			}, queryFilter...),
			"recordType": listType,
		},
		"resultsLimit": limit,
		"desiredKeys": []string{
			"resJPEGFullWidth",
			"resJPEGFullHeight",
			"resJPEGFullFileType",
			"resJPEGFullFingerprint",
			"resJPEGFullRes",
			"resJPEGLargeWidth",
			"resJPEGLargeHeight",
			"resJPEGLargeFileType",
			"resJPEGLargeFingerprint",
			"resJPEGLargeRes",
			"resJPEGMedWidth",
			"resJPEGMedHeight",
			"resJPEGMedFileType",
			"resJPEGMedFingerprint",
			"resJPEGMedRes",
			"resJPEGThumbWidth",
			"resJPEGThumbHeight",
			"resJPEGThumbFileType",
			"resJPEGThumbFingerprint",
			"resJPEGThumbRes",
			"resVidFullWidth",
			"resVidFullHeight",
			"resVidFullFileType",
			"resVidFullFingerprint",
			"resVidFullRes",
			"resVidMedWidth",
			"resVidMedHeight",
			"resVidMedFileType",
			"resVidMedFingerprint",
			"resVidMedRes",
			"resVidSmallWidth",
			"resVidSmallHeight",
			"resVidSmallFileType",
			"resVidSmallFingerprint",
			"resVidSmallRes",
			"resSidecarWidth",
			"resSidecarHeight",
			"resSidecarFileType",
			"resSidecarFingerprint",
			"resSidecarRes",
			"itemType",
			"dataClassType",
			"filenameEnc",
			"originalOrientation",
			"resOriginalWidth",
			"resOriginalHeight",
			"resOriginalFileType",
			"resOriginalFingerprint",
			"resOriginalRes",
			"resOriginalAltWidth",
			"resOriginalAltHeight",
			"resOriginalAltFileType",
			"resOriginalAltFingerprint",
			"resOriginalAltRes",
			"resOriginalVidComplWidth",
			"resOriginalVidComplHeight",
			"resOriginalVidComplFileType",
			"resOriginalVidComplFingerprint",
			"resOriginalVidComplRes",
			"isDeleted",
			"isExpunged",
			"dateExpunged",
			"remappedRef",
			"recordName",
			"recordType",
			"recordChangeTag",
			"masterRef",
			"adjustmentRenderType",
			"assetDate",
			"addedDate",
			"isFavorite",
			"isHidden",
			"orientation",
			"duration",
			"assetSubtype",
			"assetSubtypeV2",
			"assetHDRType",
			"burstFlags",
			"burstFlagsExt",
			"burstId",
			"captionEnc",
			"locationEnc",
			"locationV2Enc",
			"locationLatitude",
			"locationLongitude",
			"adjustmentType",
			"timeZoneOffset",
			"vidComplDurValue",
			"vidComplDurScale",
			"vidComplDispValue",
			"vidComplDispScale",
			"vidComplVisibilityState",
			"customRenderedValue",
			"containerId",
			"itemId",
			"position",
			"isKeyAsset",
		},
		"zoneID": map[string]any{"zoneName": "PrimarySync"},
	}

	return res
}

type getPhotosResp struct {
	Records            []*photoRecord `json:"records"`
	ContinuationMarker string         `json:"continuationMarker"`
	SyncToken          string         `json:"syncToken"`
}

type photoRecord struct {
	RecordName string `json:"recordName"`
	RecordType string `json:"recordType"`
	Fields     struct {
		ItemType                       strValue `json:"itemType,omitempty"`
		ResJPEGThumbFingerprint        strValue `json:"resJPEGThumbFingerprint,omitempty"`
		FilenameEnc                    strValue `json:"filenameEnc,omitempty"`
		ResJPEGMedRes                  urlValue `json:"resJPEGMedRes,omitempty"`
		OriginalOrientation            intValue `json:"originalOrientation,omitempty"`
		ResJPEGMedHeight               intValue `json:"resJPEGMedHeight,omitempty"`
		ResOriginalRes                 urlValue `json:"resOriginalRes,omitempty"`
		ResJPEGMedFileType             strValue `json:"resJPEGMedFileType,omitempty"`
		ResJPEGThumbHeight             intValue `json:"resJPEGThumbHeight,omitempty"`
		ResJPEGThumbWidth              intValue `json:"resJPEGThumbWidth,omitempty"`
		ResOriginalWidth               intValue `json:"resOriginalWidth,omitempty"`
		ResJPEGThumbFileType           strValue `json:"resJPEGThumbFileType,omitempty"`
		DataClassType                  intValue `json:"dataClassType,omitempty"`
		ResOriginalFingerprint         strValue `json:"resOriginalFingerprint,omitempty"`
		ResJPEGMedWidth                intValue `json:"resJPEGMedWidth,omitempty"`
		ResJPEGThumbRes                urlValue `json:"resJPEGThumbRes,omitempty"`
		ResOriginalFileType            strValue `json:"resOriginalFileType,omitempty"`
		ResOriginalHeight              intValue `json:"resOriginalHeight,omitempty"`
		ResJPEGMedFingerprint          strValue `json:"resJPEGMedFingerprint,omitempty"`
		ResVidSmallHeight              intValue `json:"resVidSmallHeight,omitempty"`
		ResOriginalVidComplFileType    strValue `json:"resOriginalVidComplFileType,omitempty"`
		ResOriginalVidComplWidth       intValue `json:"resOriginalVidComplWidth,omitempty"`
		ResVidMedFileType              strValue `json:"resVidMedFileType,omitempty"`
		ResVidMedRes                   urlValue `json:"resVidMedRes,omitempty"`
		ResVidSmallFingerprint         strValue `json:"resVidSmallFingerprint,omitempty"`
		ResVidMedWidth                 intValue `json:"resVidMedWidth,omitempty"`
		ResOriginalVidComplFingerprint strValue `json:"resOriginalVidComplFingerprint,omitempty"`
		ResVidSmallFileType            strValue `json:"resVidSmallFileType,omitempty"`
		ResVidSmallRes                 urlValue `json:"resVidSmallRes,omitempty"`
		ResOriginalVidComplRes         urlValue `json:"resOriginalVidComplRes,omitempty"`
		ResVidMedFingerprint           strValue `json:"resVidMedFingerprint,omitempty"`
		ResVidMedHeight                intValue `json:"resVidMedHeight,omitempty"`
		ResOriginalVidComplHeight      intValue `json:"resOriginalVidComplHeight,omitempty"`
		ResVidSmallWidth               intValue `json:"resVidSmallWidth,omitempty"`
		AssetDate                      intValue `json:"assetDate,omitempty"`
		Orientation                    intValue `json:"orientation,omitempty"`
		AddedDate                      intValue `json:"addedDate,omitempty"`
		AssetSubtypeV2                 intValue `json:"assetSubtypeV2,omitempty"`
		AssetHDRType                   intValue `json:"assetHDRType,omitempty"`
		TimeZoneOffset                 intValue `json:"timeZoneOffset,omitempty"`
		MasterRef                      struct {
			Value struct {
				RecordName string    `json:"recordName"`
				Action     string    `json:"action"`
				ZoneID     zoneValue `json:"zoneID"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"masterRef,omitempty"`
		AdjustmentRenderType    intValue `json:"adjustmentRenderType,omitempty"`
		VidComplDispScale       intValue `json:"vidComplDispScale,omitempty"`
		IsHidden                intValue `json:"isHidden,omitempty"`
		Duration                intValue `json:"duration,omitempty"`
		BurstFlags              intValue `json:"burstFlags,omitempty"`
		AssetSubtype            intValue `json:"assetSubtype,omitempty"`
		VidComplDurScale        intValue `json:"vidComplDurScale,omitempty"`
		VidComplDurValue        intValue `json:"vidComplDurValue,omitempty"`
		VidComplVisibilityState intValue `json:"vidComplVisibilityState,omitempty"`
		CustomRenderedValue     intValue `json:"customRenderedValue,omitempty"`
		IsFavorite              intValue `json:"isFavorite,omitempty"`
		VidComplDispValue       intValue `json:"vidComplDispValue,omitempty"`
		LocationEnc             strValue `json:"locationEnc,omitempty"`
	} `json:"fields"`
	PluginFields    struct{}       `json:"pluginFields"`
	RecordChangeTag string         `json:"recordChangeTag"`
	Created         timestampValue `json:"created"`
	Modified        timestampValue `json:"modified"`
	Deleted         bool           `json:"deleted"`
	ZoneID          zoneValue      `json:"zoneID"`
}

type intValue struct {
	Value int64  `json:"value"`
	Type  string `json:"type"` // INT64, TIMESTAMP,
}

type strValue struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type urlValue struct {
	Value urlValueVal `json:"value"`
	Type  string      `json:"type"`
}

type timestampValue struct {
	Timestamp      int64  `json:"timestamp"`
	UserRecordName string `json:"userRecordName"`
	DeviceID       string `json:"deviceID"`
}

type urlValueVal struct {
	FileChecksum      string `json:"fileChecksum"`
	Size              int    `json:"size"`
	WrappingKey       string `json:"wrappingKey"`
	ReferenceChecksum string `json:"referenceChecksum"`
	DownloadURL       string `json:"downloadURL"`
}

type zoneValue struct {
	ZoneName        string `json:"zoneName"`
	OwnerRecordName string `json:"ownerRecordName"`
	ZoneType        string `json:"zoneType"`
}
