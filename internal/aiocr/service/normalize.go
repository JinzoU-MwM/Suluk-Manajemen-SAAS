package service

func validateExtractedData(data any, docType string) []ValidationError {
	var errors []ValidationError

	extracted, ok := data.(ExtractedFields)
	if !ok {
		return nil
	}

	switch docType {
	case "ktp":
		if extracted.NIK == "" {
			errors = append(errors, ValidationError{Field: "nik", Message: "NIK tidak terdeteksi"})
		} else if len(extracted.NIK) != 16 {
			errors = append(errors, ValidationError{Field: "nik", Message: "NIK harus 16 digit", Value: extracted.NIK})
		}
		if extracted.Nama == "" {
			errors = append(errors, ValidationError{Field: "nama", Message: "Nama tidak terdeteksi"})
		}
	case "paspor":
		if extracted.NoPaspor == "" {
			errors = append(errors, ValidationError{Field: "no_paspor", Message: "Nomor paspor tidak terdeteksi"})
		}
		if extracted.NamaPaspor == "" && extracted.Nama == "" {
			errors = append(errors, ValidationError{Field: "nama", Message: "Nama tidak terdeteksi"})
		}
	case "visa":
		if extracted.NoVisa == "" {
			errors = append(errors, ValidationError{Field: "no_visa", Message: "Nomor visa tidak terdeteksi"})
		}
	}

	return errors
}

func normalizeToSiskopatuh(data any, docType string) any {
	extracted, ok := data.(ExtractedFields)
	if !ok {
		return data
	}

	normalized := map[string]any{}

	if extracted.Nama != "" {
		normalized["nama"] = extracted.Nama
	}
	// Identity number = NIK for a KTP, but the PASSPORT NUMBER for a passport
	// (a passport has no NIK). Matches template jamaah.xlsm columns D/E.
	if extracted.NIK != "" {
		normalized["no_identitas"] = extracted.NIK
		normalized["jenis_identitas"] = "KTP"
	} else if extracted.NoPaspor != "" {
		normalized["no_identitas"] = extracted.NoPaspor
		normalized["jenis_identitas"] = "Paspor"
	}
	if extracted.NoPaspor != "" {
		normalized["no_paspor"] = extracted.NoPaspor
	}
	if extracted.NamaPaspor != "" {
		normalized["nama_paspor"] = extracted.NamaPaspor
	} else if extracted.Nama != "" {
		normalized["nama_paspor"] = extracted.Nama
	}
	if extracted.TempatLahir != "" {
		normalized["tempat_lahir"] = extracted.TempatLahir
	}
	if extracted.TanggalLahir != "" {
		normalized["tanggal_lahir"] = extracted.TanggalLahir
	}
	if extracted.JenisKelamin != "" {
		normalized["gender"] = extracted.JenisKelamin
	}
	if extracted.Alamat != "" {
		normalized["alamat"] = extracted.Alamat
	}
	if extracted.Provinsi != "" {
		normalized["provinsi"] = extracted.Provinsi
	}
	if extracted.Kabupaten != "" {
		normalized["kabupaten"] = extracted.Kabupaten
	}
	if extracted.Kecamatan != "" {
		normalized["kecamatan"] = extracted.Kecamatan
	}
	if extracted.Kelurahan != "" {
		normalized["kelurahan"] = extracted.Kelurahan
	}
	if extracted.NoTelepon != "" {
		normalized["no_telepon"] = extracted.NoTelepon
	}
	if extracted.NoHP != "" {
		normalized["no_hp"] = extracted.NoHP
	}
	if extracted.Kewarganegaraan != "" {
		normalized["kewarganegaraan"] = extracted.Kewarganegaraan
	}
	if extracted.GolonganDarah != "" {
		normalized["golongan_darah"] = extracted.GolonganDarah
	}
	if extracted.Pendidikan != "" {
		normalized["pendidikan"] = extracted.Pendidikan
	}
	if extracted.Pekerjaan != "" {
		normalized["pekerjaan"] = extracted.Pekerjaan
	}
	if extracted.StatusPerkawinan != "" {
		normalized["status_pernikahan"] = mapStatusNikah(extracted.StatusPerkawinan)
	}
	if extracted.Agama != "" {
		normalized["agama"] = extracted.Agama
	}

	if extracted.TanggalPaspor != "" {
		normalized["tanggal_paspor"] = extracted.TanggalPaspor
	}
	if extracted.TanggalExpired != "" {
		normalized["tanggal_expired_paspor"] = extracted.TanggalExpired
	}
	if extracted.KotaPaspor != "" {
		normalized["kota_paspor"] = extracted.KotaPaspor
	}

	if extracted.ProviderVisa != "" {
		normalized["provider_visa"] = extracted.ProviderVisa
	}
	if extracted.NoVisa != "" {
		normalized["no_visa"] = extracted.NoVisa
	}
	if extracted.TanggalVisa != "" {
		normalized["tanggal_visa"] = extracted.TanggalVisa
	}
	if extracted.TanggalVisaAkhir != "" {
		normalized["tanggal_visa_akhir"] = extracted.TanggalVisaAkhir
	}

	normalized["source_doc_type"] = docType
	normalized["siskopatuh_version"] = "2.0"

	return normalized
}
