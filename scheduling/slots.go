package scheduling

import (
	"github.com/keremk/challenge-bot/models"
)

type SlotInfo struct {
	Slot       models.Slot
	IsSelected bool
}

func SlotsForWeek(week, year int, reviewer models.Reviewer, challenge models.ChallengeSetup) []SlotInfo {
	slotIndex := models.SlotIndex(week, year)

	slots := make([]SlotInfo, 0, len(challenge.Slots))
	availableSlots := reviewer.Availability[slotIndex]
	if availableSlots == nil || len(availableSlots) == 0 {
		for _, slot := range challenge.Slots {
			slots = append(slots, SlotInfo{
				Slot:       slot,
				IsSelected: isSelected(slot, reviewer.Availability["General"]),
			})
		}
	} else {
		for _, slot := range challenge.Slots {
			slots = append(slots, SlotInfo{
				Slot:       slot,
				IsSelected: isSelected(slot, availableSlots),
			})
		}
	}

	return slots
}

func isSelected(slot models.Slot, availableSlots []string) bool {
	for _, slotID := range availableSlots {
		if slotID == slot.ID {
			return true
		}
	}
	return false
}
