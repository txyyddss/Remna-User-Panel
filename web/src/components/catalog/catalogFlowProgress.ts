const completedIcon = 'i-ph-check-bold'

export function indicatorIcon(value: number, currentStep: number, icon: string): string {
  return value < currentStep ? completedIcon : icon
}
