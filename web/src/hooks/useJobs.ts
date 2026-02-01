import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../services/api';

export interface JobStats {
    total_jobs: number;
    pending_jobs: number;
    running_jobs: number;
    failed_jobs: number;
    completed_jobs: number;
}

export interface Job {
    id: string;
    type: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    created_at: string;
    retry_count: number;
    payload?: string;
    error_message?: string;
}

export interface JobsResponse {
    jobs: Job[];
    meta: {
        current_page: number;
        total_pages: number;
        total_records: number;
    };
}

export interface CreateJobPayload {
    type: string;
    payload: string;
}

// Hooks

interface RawJobStats {
    totalJobs: string;
    pendingJobs: string;
    runningJobs: string;
    failedJobs: string;
    completedJobs: string;
}

interface RawJob {
    jobId: string;
    type: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    createdAt: string;
    retryCount: string;
    payload?: string;
    errorMessage?: string;
}

interface RawJobsResponse {
    jobs: RawJob[];
    meta: {
        currentPage: number;
        totalPages: number;
        totalRecords: string;
    };
}

const mapJob = (raw: RawJob): Job => ({
    id: raw.jobId,
    type: raw.type,
    status: raw.status,
    created_at: raw.createdAt,
    retry_count: parseInt(raw.retryCount, 10),
    payload: raw.payload,
    error_message: raw.errorMessage,
});

export const useJobStats = () => {
    return useQuery({
        queryKey: ['jobStats'],
        queryFn: async () => {
            const { data } = await api.get<RawJobStats>('/stats');
            return {
                total_jobs: parseInt(data.totalJobs, 10),
                pending_jobs: parseInt(data.pendingJobs, 10),
                running_jobs: parseInt(data.runningJobs, 10),
                failed_jobs: parseInt(data.failedJobs, 10),
                completed_jobs: parseInt(data.completedJobs, 10),
            };
        },
        refetchInterval: 2000,
    });
};

export const useJobs = (limit: number = 20, offset: number = 0) => {
    return useQuery({
        queryKey: ['jobs', limit, offset],
        queryFn: async () => {
            const { data } = await api.get<RawJobsResponse>(`/jobs`, {
                params: { limit, offset },
            });
            return {
                jobs: data.jobs.map(mapJob),
                meta: {
                    current_page: data.meta.currentPage,
                    total_pages: data.meta.totalPages,
                    total_records: parseInt(data.meta.totalRecords, 10),
                },
            };
        },
        refetchInterval: 2000,
    });
};

export const useJob = (id: string) => {
    return useQuery({
        queryKey: ['job', id],
        queryFn: async () => {
            const { data } = await api.get<RawJob>(`/jobs/${id}`);
            return mapJob(data);
        },
        enabled: !!id,
    });
};

export const useCreateJob = () => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (newJob: CreateJobPayload) => {
            const { data } = await api.post('/jobs', newJob);
            return data;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['jobs'] });
            queryClient.invalidateQueries({ queryKey: ['jobStats'] });
        },
    });
};
