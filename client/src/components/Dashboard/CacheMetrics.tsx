import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ResponsiveLine } from '@nivo/line';
import subMinutes from 'date-fns/sub_minutes';
import dateFormat from 'date-fns/format';
import round from 'lodash/round';
import Card from '../ui/Card';
import { STATUS_COLORS } from '../../helpers/constants';
import './MetricsCards.css';
import '../ui/Line.css';

interface CacheMetricsData {
    cache_enabled: boolean;
    cache_hit_rate: number;
    cache_size: number;
    total_queries: number;
    cache_hits: number;
    cache_misses: number;
    history: number[];
}

const REFRESH_INTERVAL = 30000; // 30 seconds

const CacheMetrics = () => {
    const { t } = useTranslation();
    const [data, setData] = useState<CacheMetricsData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchMetrics = async () => {
        try {
            const response = await fetch('/control/cache_metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const result = await response.json();
            setData(result);
            setError(null);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Unknown error';
            console.error('Failed to fetch cache metrics:', errorMessage);
            setError(t('failed_to_load_cache_metrics'));
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
            <Card type="card--full" bodyType="card-wrap">
                <div className="card-body-stats">
                    <div className="card-value card-value-stats text-gray">--</div>
                    <div className="card-title-stats">{t('dns_cache_hit_rate')}</div>
                </div>
                <div className="card-chart-bg">
                    <div className="card-disabled-overlay">{t('loading')}</div>
                </div>
            </Card>
        );
    }

    if (error) {
        return (
            <Card type="card--full" bodyType="card-wrap">
                <div className="card-body-stats">
                    <div className="card-value card-value-stats text-gray">--</div>
                    <div className="card-title-stats">{t('dns_cache_hit_rate')}</div>
                </div>
                <div className="card-chart-bg">
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

    if (!data.cache_enabled) {
        return (
            <Card type="card--full" bodyType="card-wrap">
                <div className="card-body-stats">
                    <div className="card-value card-value-stats text-gray">--</div>
                    <div className="card-title-stats">{t('dns_cache_hit_rate')}</div>
                </div>
                <div className="card-chart-bg">
                    <div className="card-disabled-overlay">{t('cache_disabled')}</div>
                </div>
            </Card>
        );
    }

    const hitRate = data.cache_hit_rate.toFixed(1);
    
    // Use backend-provided history data (60 minutes)
    const lineData = [
        {
            id: 'cacheHitRate',
            data: data.history.map((value, index) => ({
                x: index,
                y: value,
            })),
        },
    ];

    return (
        <Card type="card--full" bodyType="card-wrap">
            <div className="card-body-stats">
                <div className="card-value card-value-stats text-blue">{hitRate}%</div>
                <div className="card-title-stats">{t('dns_cache_hit_rate')}</div>
            </div>
            <div className="card-chart-bg">
                <ResponsiveLine
                    enableArea
                    animate
                    enableSlices="x"
                    curve="linear"
                    colors={[STATUS_COLORS.blue]}
                    data={lineData}
                    theme={{
                        crosshair: {
                            line: {
                                stroke: 'currentColor',
                                strokeWidth: 1,
                                strokeOpacity: 0.5,
                            },
                        },
                    }}
                    xScale={{
                        type: 'linear',
                        min: 0,
                        max: 59,
                    }}
                    yScale={{
                        type: 'linear',
                        min: 0,
                        max: 100,
                    }}
                    crosshairType="x"
                    axisLeft={null}
                    axisBottom={null}
                    enableGridX={false}
                    enableGridY={false}
                    enablePoints={false}
                    xFormat={(x: number) => {
                        // x is the minute index (0-59)
                        // Calculate the actual time: current time - (59 - x) minutes
                        const minutesAgo = 59 - x;
                        const time = subMinutes(Date.now(), minutesAgo);
                        return dateFormat(time, 'HH:mm');
                    }}
                    yFormat={(y: number) => `${round(y, 1)}%`}
                    sliceTooltip={(slice) => {
                        const { xFormatted, yFormatted } = slice.slice.points[0].data;
                        return (
                            <div className="line__tooltip">
                                <span className="line__tooltip-text">
                                    <strong>{yFormatted}</strong>
                                    <br />
                                    <small>{xFormatted}</small>
                                </span>
                            </div>
                        );
                    }}
                />
            </div>
        </Card>
    );
};

export default CacheMetrics;
