/**
 * Time utility functions for the chat application
 */

/**
 * Format a timestamp according to the user's local timezone
 * 
 * @param timestamp - Unix timestamp in milliseconds or Date object
 * @param format - Optional format string ('short', 'long', 'time', 'date', 'datetime')
 * @param timezone - Optional timezone (defaults to user's local timezone)
 * @returns Formatted date/time string
 */
export function formatTimestamp(
  timestamp: number | Date,
  format: 'short' | 'long' | 'time' | 'date' | 'datetime' = 'short',
  timezone?: string
): string {
  const date = timestamp instanceof Date ? timestamp : new Date(timestamp);
  
  // Default to user's local timezone if not specified
  const tz = timezone || Intl.DateTimeFormat().resolvedOptions().timeZone;
  
  // Define format options based on the requested format
  let options: Intl.DateTimeFormatOptions;
  
  switch (format) {
    case 'short':
      options = { 
        timeZone: tz,
        hour: 'numeric', 
        minute: '2-digit'
      };
      
      // Add the date if it's not today
      if (!isToday(date)) {
        options.month = 'short';
        options.day = 'numeric';
      }
      
      // Add the year if it's not the current year
      if (!isCurrentYear(date)) {
        options.year = 'numeric';
      }
      break;
      
    case 'long':
      options = { 
        timeZone: tz,
        weekday: 'long',
        year: 'numeric', 
        month: 'long', 
        day: 'numeric',
        hour: 'numeric', 
        minute: '2-digit',
        second: '2-digit'
      };
      break;
      
    case 'time':
      options = { 
        timeZone: tz,
        hour: 'numeric', 
        minute: '2-digit'
      };
      break;
      
    case 'date':
      options = { 
        timeZone: tz,
        year: 'numeric', 
        month: 'short', 
        day: 'numeric'
      };
      break;
      
    case 'datetime':
      options = { 
        timeZone: tz,
        year: 'numeric', 
        month: 'short', 
        day: 'numeric',
        hour: 'numeric', 
        minute: '2-digit'
      };
      break;
  }
  
  return new Intl.DateTimeFormat('en', options).format(date);
}

/**
 * Check if a date is today in the user's timezone
 * 
 * @param date - Date to check
 * @param timezone - Optional timezone (defaults to user's local timezone)
 * @returns True if the date is today
 */
export function isToday(date: Date, timezone?: string): boolean {
  const tz = timezone || Intl.DateTimeFormat().resolvedOptions().timeZone;
  const now = new Date();
  
  const dateOptions: Intl.DateTimeFormatOptions = { 
    timeZone: tz,
    year: 'numeric', 
    month: 'numeric', 
    day: 'numeric'
  };
  
  const formatter = new Intl.DateTimeFormat('en', dateOptions);
  
  return formatter.format(date) === formatter.format(now);
}

/**
 * Check if a date is in the current year in the user's timezone
 * 
 * @param date - Date to check
 * @param timezone - Optional timezone (defaults to user's local timezone)
 * @returns True if the date is in the current year
 */
export function isCurrentYear(date: Date, timezone?: string): boolean {
  const tz = timezone || Intl.DateTimeFormat().resolvedOptions().timeZone;
  const now = new Date();
  
  const yearOptions: Intl.DateTimeFormatOptions = { 
    timeZone: tz,
    year: 'numeric'
  };
  
  const formatter = new Intl.DateTimeFormat('en', yearOptions);
  
  return formatter.format(date) === formatter.format(now);
}

/**
 * Convert a date to the user's timezone
 * 
 * @param date - Date to convert
 * @param timezone - Optional timezone (defaults to user's local timezone)
 * @returns Date object adjusted to the user's timezone
 */
export function convertToUserTimezone(date: Date, timezone?: string): Date {
  // This function doesn't actually modify the date object,
  // since Date objects are always in the local timezone of the browser.
  // It's included for API completeness and potential future enhancements.
  return date;
}

/**
 * Get the user's current timezone
 * 
 * @returns The user's timezone
 */
export function getUserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * Format a timestamp as a relative time (e.g., "5 minutes ago")
 * 
 * @param timestamp - Unix timestamp in milliseconds or Date object
 * @returns Formatted relative time string
 */
export function formatRelativeTime(timestamp: number | Date): string {
  const date = timestamp instanceof Date ? timestamp : new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  
  if (diffSecs < 60) {
    return 'just now';
  } else if (diffMins < 60) {
    return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  } else if (diffDays < 7) {
    return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  } else {
    return formatTimestamp(date, 'date');
  }
} 