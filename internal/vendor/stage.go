package vendor

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
