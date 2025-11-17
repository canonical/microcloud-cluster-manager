export const getDefaultExpiryDate = (): string => {
  // The datetime selector only selects a default value in the Simplified ISO8601 DateTime format, however the backend only accepts values in the full datetime formattedDate. This function provides the default date in the correct format.

  const now = new Date();
  now.setHours(now.getHours() + 24);

  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0"); // Months are zero-based
  const day = String(now.getDate()).padStart(2, "0");
  const formattedHours = String(now.getHours()).padStart(2, "0");
  const formattedMinutes = String(now.getMinutes()).padStart(2, "0");

  // Combine the components into the desired format
  return `${year}-${month}-${day}T${formattedHours}:${formattedMinutes}`;
};
