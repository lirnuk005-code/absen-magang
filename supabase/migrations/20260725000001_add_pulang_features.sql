-- Migration: Add type and early_reason columns to attendance_logs table
ALTER TABLE public.attendance_logs 
ADD COLUMN IF NOT EXISTS type TEXT DEFAULT 'DATANG',
ADD COLUMN IF NOT EXISTS early_reason TEXT DEFAULT NULL;
