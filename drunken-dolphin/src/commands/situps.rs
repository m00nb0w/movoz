use crate::error::AppError;
use crate::personal::Tracker;

pub fn execute(tracker: &mut Tracker, count: u32, date: &str) -> Result<(), AppError> {
    tracker.record_situps(count, date)
}
