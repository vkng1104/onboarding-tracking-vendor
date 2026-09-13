package vendor

func isValidStage(stage Stage) bool {
	switch stage {
	case StageContractSent,
		StageContractSigned,
		StageKYCDocsReceived,
		StageKYCVerified,
		StageActive:
		return true
	default:
		return false
	}
}

func nextStage(current Stage) *Stage {
	var next Stage
	switch current {
	case StageContractSent:
		next = StageContractSigned
	case StageContractSigned:
		next = StageKYCDocsReceived
	case StageKYCDocsReceived:
		next = StageKYCVerified
	case StageKYCVerified:
		next = StageActive
	case StageActive:
		return nil
	default:
		return nil
	}
	return &next
}
