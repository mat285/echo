package logger

// MaybeTrigger triggers an event if the logger is set.
func MaybeTrigger(log Triggerable, e Event) {
	if log == nil {
		return
	}
	log.Trigger(e)
}

// MaybeSyncTrigger triggers an event if the logger is set.
func MaybeSyncTrigger(log SyncTriggerable, e Event) {
	if log == nil {
		return
	}
	log.SyncTrigger(e)
}

// MaybeInfof triggers Infof if the logger is set.
func MaybeInfof(log InfofReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Infof(format, args...)
}

// MaybeSyncInfof triggers SyncInfof if the logger is set.
func MaybeSyncInfof(log SyncInfofReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.SyncInfof(format, args...)
}

// MaybeDebugf triggers Debugf if the logger is set.
func MaybeDebugf(log DebugfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Debugf(format, args...)
}

// MaybeSyncDebugf triggers SyncDebugf if the logger is set.
func MaybeSyncDebugf(log SyncDebugfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.SyncDebugf(format, args...)
}

// MaybeSillyf triggers Sillyf if the logger is set.
func MaybeSillyf(log SillyfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Sillyf(format, args...)
}

// MaybeSyncSillyf triggers SyncSillyf if the logger is set.
func MaybeSyncSillyf(log SyncSillyfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.SyncSillyf(format, args...)
}

// MaybeWarningf triggers Warningf if the logger is set.
func MaybeWarningf(log WarningfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Warningf(format, args...)
}

// MaybeSyncWarningf triggers SyncWarningf if the logger is set.
func MaybeSyncWarningf(log SyncWarningfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.SyncWarningf(format, args...)
}

// MaybeWarning triggers Warning if the logger is set.
func MaybeWarning(log WarningReceiver, err error) {
	if log == nil || err == nil {
		return
	}
	log.Warning(err)
}

// MaybeSyncWarning triggers SyncWarning if the logger is set.
func MaybeSyncWarning(log SyncWarningReceiver, err error) {
	if log == nil || err == nil {
		return
	}
	log.SyncWarning(err)
}

// MaybeErrorf triggers Errorf if the logger is set.
func MaybeErrorf(log ErrorfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Errorf(format, args...)
}

// MaybeSyncErrorf triggers SyncErrorf if the logger is set.
func MaybeSyncErrorf(log SyncErrorffReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.SyncErrorf(format, args...)
}

// MaybeError triggers Error if the logger is set.
func MaybeError(log ErrorReceiver, err error) {
	if log == nil || err == nil {
		return
	}
	log.Error(err)
}

// MaybeSyncError triggers SyncError if the logger is set.
func MaybeSyncError(log SyncErrorReceiver, err error) {
	if log == nil || err == nil {
		return
	}
	log.SyncError(err)
}

// MaybeFatalf triggers Fatalf if the logger is set.
func MaybeFatalf(log FatalfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Fatalf(format, args...)
}

// MaybeSyncFatalf triggers SyncFatalf if the logger is set.
func MaybeSyncFatalf(log SyncFatalfReceiver, format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.SyncFatalf(format, args...)
}

// MaybeFatal triggers Fatal if the logger is set.
func MaybeFatal(log FatalReceiver, err error) {
	if log == nil || err == nil {
		return
	}
	log.Fatal(err)
}

// MaybeSyncFatal triggers SyncFatal if the logger is set.
func MaybeSyncFatal(log SyncFatalReceiver, err error) {
	if log == nil || err == nil {
		return
	}
	log.SyncFatal(err)
}
