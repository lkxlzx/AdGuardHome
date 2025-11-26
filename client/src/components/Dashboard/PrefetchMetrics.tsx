import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import Card from '../ui/Card';
import { formatNumber } from '../../helpers/helpers';
import './MetricsCards.css';

interface PrefetchMetricsData {
    prefetch_enabled: boolean;
    prefetch_status: string;
    prefetch_hot_domains: number;
    prefetch_completed: number;
    prefetch_failed: number;
    prefetch_success_rate: number;
    prefetch_queue_size: number;
    last_prefetch_time: string;
}

const REFRESH_INTERVAL = 30000; // 30 seconds

const PrefetchMetrics = () => {
    const { t } = useTranslation();
    const [data, setData] = useState<PrefetchMetricsData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchMetrics = async () => {
        try {
            const response = await fetch('/control/prefetch_metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const result = await response.json();
            setData(result);
            setError(null);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Unknown error';
            console.error('Failed to fetch prefetch metrics:', errorMessage);
            setError(t('failed_to_load_prefetch_metrics'));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchMetrics();
        const interval = setInterval(fetchMetrics, REFRESH_INTERVAL);
        return () => clearInterval(interval);
    }, []);

    if (loading) {
        return (
            <Card type="card--full" bodyType="card-body">
                <div className="prefetch-card">
                    <div className="prefetch-card__header">
                        <svg className="icons icon--24 icon--gray">
                            <use xlinkHref="#refresh" />
                        </svg>
                        <span className="prefetch-card__title">{t('prefetch_status')}</span>
                    </div>
                    <div className="prefetch-card__disabled">
                        <span className="badge badge-secondary">{t('loading')}</span>
                    </div>
                </div>
            </Card>
        );
    }

    if (error) {
        return (
            <Card type="card--full" bodyType="card-body">
                <div className="prefetch-card">
                    <div className="prefetch-card__header">
                        <svg className="icons icon--24 icon--red">
                            <use xlinkHref="#refresh" />
                        </svg>
                        <span className="prefetch-card__title">{t('prefetch_status')}</span>
                    </div>
                    <div className="alert alert-danger" style={{ margin: '10px' }}>
                        {error}
                    </div>
                </div>
            </Card>
        );
    }

    if (!data) {
        return null;
    }

    if (!data.prefetch_enabled) {
        return (
            <Card type="card--full" bodyType="card-body">
                <div className="prefetch-card">
                    <div className="prefetch-card__header">
                        <svg className="icons icon--24 icon--gray">
                            <use xlinkHref="#refresh" />
                        </svg>
                        <span className="prefetch-card__title">{t('prefetch_status')}</span>
                    </div>
                    <div className="prefetch-card__disabled">
                        <span className="badge badge-secondary">{t('disabled')}</span>
                    </div>
                </div>
            </Card>
        );
    }

    const statusColor = data.prefetch_status === 'active' ? 'success' : 'secondary';
    const statusText = data.prefetch_status === 'active' ? t('prefetch_active') : t('prefetch_idle');

    return (
        <Card type="card--full" bodyType="card-body">
            <div className="prefetch-card">
                <div className="prefetch-card__header">
                    <svg className="icons icon--24 icon--green">
                        <use xlinkHref="#refresh" />
                    </svg>
                    <span className="prefetch-card__title">{t('prefetch_status')}</span>
                    <span className={`badge badge-${statusColor} ml-auto`}>{statusText}</span>
                </div>

                <div className="prefetch-card__stats">
                    <div className="prefetch-stat">
                        <div className="prefetch-stat__value text-success">
                            {data.prefetch_success_rate.toFixed(1)}%
                        </div>
                        <div className="prefetch-stat__label">{t('success_rate')}</div>
                    </div>

                    <div className="prefetch-stat">
                        <div className="prefetch-stat__value">{formatNumber(data.prefetch_hot_domains)}</div>
                        <div className="prefetch-stat__label">{t('hot_domains')}</div>
                    </div>

                    <div className="prefetch-stat">
                        <div className="prefetch-stat__value">{formatNumber(data.prefetch_queue_size)}</div>
                        <div className="prefetch-stat__label">{t('queue_size')}</div>
                    </div>
                </div>

                <div className="prefetch-card__details">
                    <div className="prefetch-detail">
                        <span className="prefetch-detail__label">{t('completed')}:</span>
                        <span className="prefetch-detail__value text-success">
                            {formatNumber(data.prefetch_completed)}
                        </span>
                    </div>
                    <div className="prefetch-detail">
                        <span className="prefetch-detail__label">{t('failed')}:</span>
                        <span className="prefetch-detail__value text-danger">
                            {formatNumber(data.prefetch_failed)}
                        </span>
                    </div>
                    {data.last_prefetch_time && (
                        <div className="prefetch-detail">
                            <span className="prefetch-detail__label">{t('last_prefetch')}:</span>
                            <span className="prefetch-detail__value text-muted">
                                {new Date(data.last_prefetch_time).toLocaleString()}
                            </span>
                        </div>
                    )}
                </div>
            </div>
        </Card>
    );
};

export default PrefetchMetrics;
