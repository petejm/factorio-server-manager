import React, {useEffect, useState, useCallback} from "react";
import Panel from "../components/Panel";
import Button from "../components/Button";
import server from "../../api/resources/server";
import savesResource from "../../api/resources/saves";
import {useForm} from "react-hook-form";
import Select from "../components/Select";
import Input from "../components/Input";
import Error from "../components/Error";

const Controls = ({serverStatus}) => {

    const factorioVersion = serverStatus.fac_version ? serverStatus.fac_version : 'Unknown';
    const [saves, setSaves] = useState([]);
    const [isDisabled, setIsDisabled] = useState(true);
    const [isStopping, setIsStopping] = useState(false);
    const [isStarting, setIsStarting] = useState(false);
    const [isKilling, setIsKilling] = useState(false);
    const [gameInfo, setGameInfo] = useState(null);
    const [isLoadingGameInfo, setIsLoadingGameInfo] = useState(false);

    const { handleSubmit, reset, register, formState: {errors} } = useForm();

    const startServer = async (data) => {
        setIsStarting(true);
        try {
            await server.start(data.ip, parseInt(data.port), data.save);
        } finally {
            setIsStarting(false);
        }
    }

    const stopServer = async () => {
        setIsStopping(true);
        try {
            await server.stop();
        } finally {
            setIsStopping(false);
        }
    }

    const killServer = async () => {
        setIsKilling(true);
        try {
            await server.kill();
        } finally {
            setIsKilling(false);
        }
    }

    const fetchGameInfo = useCallback(async () => {
        if (serverStatus.status !== 'running') return;
        setIsLoadingGameInfo(true);
        try {
            const info = await server.gameInfo();
            setGameInfo(info);
        } catch (err) {
            console.error('Failed to fetch game info:', err);
        } finally {
            setIsLoadingGameInfo(false);
        }
    }, [serverStatus.status]);

    useEffect(() => {
        savesResource.list(true)
            .then(res => {
                setSaves(res);
                if (res.length > 0) {
                    setIsDisabled(false);
                }
                reset();
            });
    }, [])

    // Fetch game info when server starts running
    useEffect(() => {
        if (serverStatus.status === 'running') {
            fetchGameInfo();
        } else {
            setGameInfo(null);
        }
    }, [serverStatus.status, fetchGameInfo])

    return (
        <>
        <form onSubmit={handleSubmit(startServer)}>
        <Panel
            title="Server Status"
            content={
                <div className="lg:flex">
                    { serverStatus.status === 'running'
                        ? <>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">Status</div>
                                <div>{serverStatus.status === 'running' ? 'Running' : 'Stopped'}</div>
                            </div>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">IP</div>
                                <div>{serverStatus.bindip}</div>
                            </div>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">Port</div>
                                <div>{serverStatus.port}</div>
                            </div>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">Factorio Version</div>
                                <div>{factorioVersion}</div>
                            </div>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">Save</div>
                                <div>{serverStatus.savefile}</div>
                            </div>
                        </>
                        : <>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">Status</div>
                                <div>{serverStatus.status === 'running' ? 'Running' : 'Stopped'}</div>
                            </div>
                            <div className="lg:w-1/5 mb-2 mr-0 lg:mr-4">
                                <div className="font-bold">IP</div>
                                <Input
                                    defaultValue={"0.0.0.0"}
                                    disabled={isDisabled}
                                    register={register('ip',{required: true, pattern: /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/})}
                                />
                                <Error error={errors.ip} message="IP is required and must be valid."/>
                            </div>
                            <div className="lg:w-1/5 mb-2 mr-0 lg:mr-4">
                                <div className="font-bold">Port</div>
                                <Input
                                    type="number"
                                    min={1}
                                    max={65535}
                                    defaultValue={"34197"}
                                    disabled={isDisabled}
                                    register={register('port',{required: true, min: 1, max: 65535})}
                                />
                                <Error error={errors.port} message="Port is required within range 1-65535"/>
                            </div>
                            <div className="lg:w-1/5 mb-2 mr-0 lg:mr-4">
                                <div className="font-bold">Factorio Version</div>
                                <div>{factorioVersion}</div>
                            </div>
                            <div className="lg:w-1/5 mb-2">
                                <div className="font-bold">Save</div>
                                <div className="relative">
                                    <Select
                                        register={register('save',{required: true})}
                                        defaultValue={saves.find((save) => save.name.startsWith('Load Latest'))?.name}
                                        disabled={isDisabled}
                                        options={saves.map(save => new Object({
                                            value: save.name,
                                            name: save.name
                                        }))}
                                    />
                                    <Error error={errors.save} message="Save is required and must be valid."/>
                                </div>
                            </div>
                        </>
                    }
                </div>
            }
            actions={
                <div className="md:flex">
                    {serverStatus.status === 'running'
                        ? <>
                            <Button onClick={stopServer} isLoading={isStopping} isDisabled={isKilling} size="sm" className="w-full md:w-auto mb-2 md:mb-0 md:mr-2" type="default">Save & Stop Server</Button>
                            <Button onClick={killServer} isLoading={isKilling} isDisabled={isStopping} size="sm" type="danger" className="w-full md:w-auto">Kill Server</Button>
                        </>
                        : <Button isSubmit={true} isDisabled={isDisabled} isLoading={isStarting} size="sm" type="success" className="w-full md:w-auto">Start Server</Button>
                    }
                </div>
            }
        />
        </form>

        {/* Game Info Panel - only shown when server is running */}
        {serverStatus.status === 'running' && (
            <Panel
                title="Game Info"
                content={
                    <div className="lg:flex">
                        <div className="lg:w-1/4 mb-2">
                            <div className="font-bold">Players Online</div>
                            <div>{isLoadingGameInfo ? 'Loading...' : (gameInfo?.player_count || 'N/A')}</div>
                        </div>
                        <div className="lg:w-1/4 mb-2">
                            <div className="font-bold">Game Time</div>
                            <div>{isLoadingGameInfo ? 'Loading...' : (gameInfo?.game_time || 'N/A')}</div>
                        </div>
                        <div className="lg:w-1/4 mb-2">
                            <div className="font-bold">Evolution</div>
                            <div>{isLoadingGameInfo ? 'Loading...' : (gameInfo?.evolution || 'N/A')}</div>
                        </div>
                        <div className="lg:w-1/4 mb-2">
                            <div className="font-bold">Map Seed</div>
                            <div className="text-sm break-all">{isLoadingGameInfo ? 'Loading...' : (gameInfo?.seed || 'N/A')}</div>
                        </div>
                    </div>
                }
                actions={
                    <Button
                        onClick={fetchGameInfo}
                        isLoading={isLoadingGameInfo}
                        size="sm"
                        type="default"
                        className="w-full md:w-auto"
                    >
                        Refresh
                    </Button>
                }
            />
        )}
    </>
    )
};

export default Controls;
