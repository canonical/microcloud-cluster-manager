export const getLastSeenLabel = (minutesSinceLastSeen: number): string => {
  if (minutesSinceLastSeen <= 1) {
    return "seen 1 minute ago";
  }

  if (minutesSinceLastSeen < 5) {
    return `seen ${minutesSinceLastSeen} minutes ago`;
  }

  const hoursSinceLastSeen = Math.floor(minutesSinceLastSeen / 60);

  if (hoursSinceLastSeen < 1) {
    //Displayed for "Last Seen"'s of 5-59 Minutes
    return "seen in last hour";
  }

  if (hoursSinceLastSeen < 2) {
    return "seen 1 hour ago";
  }

  return `seen ${hoursSinceLastSeen} hours ago`;
};
