import React, {useEffect, useState} from "react";
import Panel from "../components/Panel";
import Button from "../components/Button";
import Select from "../components/Select";
import backupsResource from "../../api/resources/backups";
import savesResource from "../../api/resources/saves";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faDownload, faTrashAlt, faUndo, faCheckCircle} from "@fortawesome/free-solid-svg-icons";
import {useForm} from "react-hook-form";

const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const Backups = ({serverStatus}) => {
    const [backups, setBackups] = useState([]);
    const [saves, setSaves] = useState([]);
    const [isCreating, setIsCreating] = useState(false);
    const [isRestoring, setIsRestoring] = useState(null);
    const [isDeleting, setIsDeleting] = useState(null);
    const [isVerifying, setIsVerifying] = useState(null);
    const [verifyResult, setVerifyResult] = useState(null);

    const {handleSubmit, register, reset} = useForm();

    const updateBackupList = () => {
        backupsResource.list()
            .then(res => {
                if (res) {
                    setBackups(res);
                }
            });
    };

    const updateSavesList = () => {
        savesResource.list()
            .then(res => {
                if (res) {
                    setSaves(res);
                }
            });
    };

    useEffect(() => {
        updateBackupList();
        updateSavesList();
    }, []);

    const createBackup = async (data) => {
        setIsCreating(true);
        try {
            const res = await backupsResource.create(data.save);
            if (res) {
                updateBackupList();
                reset();
            }
        } finally {
            setIsCreating(false);
        }
    };

    const restoreBackup = async (backup) => {
        if (serverStatus.status === 'running') {
            alert('Cannot restore backup while server is running. Please stop the server first.');
            return;
        }
        if (!window.confirm(`Are you sure you want to restore backup "${backup.filename}"? This will overwrite the current save file.`)) {
            return;
        }
        setIsRestoring(backup.id);
        try {
            const res = await backupsResource.restore(backup.id);
            if (res) {
                alert('Backup restored successfully.');
            }
        } finally {
            setIsRestoring(null);
        }
    };

    const deleteBackup = async (backup) => {
        if (!window.confirm(`Are you sure you want to delete backup "${backup.filename}"?`)) {
            return;
        }
        setIsDeleting(backup.id);
        try {
            const res = await backupsResource.delete(backup.id);
            if (res) {
                updateBackupList();
            }
        } finally {
            setIsDeleting(null);
        }
    };

    const verifyBackup = async (backup) => {
        setIsVerifying(backup.id);
        setVerifyResult(null);
        try {
            const res = await backupsResource.verify(backup.id);
            if (res) {
                setVerifyResult({id: backup.id, ...res});
                alert(`Checksum: ${res.checksum}\nValid: ${res.valid ? 'Yes' : 'No'}`);
            }
        } finally {
            setIsVerifying(null);
        }
    };

    const getStatusBadge = (status) => {
        const statusColors = {
            'valid': 'bg-green text-black',
            'pending': 'bg-yellow-500 text-black',
            'corrupted': 'bg-red text-white',
            'unknown': 'bg-gray-light text-black'
        };
        const colorClass = statusColors[status] || statusColors['unknown'];
        return (
            <span className={`px-2 py-1 rounded text-xs font-bold ${colorClass}`}>
                {status || 'unknown'}
            </span>
        );
    };

    const getTypeBadge = (type) => {
        const typeColors = {
            'manual': 'bg-blue-500 text-white',
            'automated': 'bg-purple-500 text-white'
        };
        const colorClass = typeColors[type] || 'bg-gray-light text-black';
        return (
            <span className={`px-2 py-1 rounded text-xs font-bold ${colorClass}`}>
                {type || 'manual'}
            </span>
        );
    };

    return (
        <>
            <Panel
                title="Create Backup"
                className="mb-6"
                content={
                    <form onSubmit={handleSubmit(createBackup)} className="flex flex-col md:flex-row md:items-end gap-4">
                        <div className="flex-grow">
                            <label className="block font-bold mb-2">Select Save File</label>
                            <Select
                                register={register('save', {required: true})}
                                disabled={saves.length === 0}
                                options={saves.map(save => ({
                                    value: save.name,
                                    name: save.name
                                }))}
                            />
                        </div>
                        <div>
                            <Button
                                isSubmit={true}
                                isDisabled={saves.length === 0}
                                isLoading={isCreating}
                                size="sm"
                                type="success"
                            >
                                Create Backup
                            </Button>
                        </div>
                    </form>
                }
            />

            <Panel
                title="Backups"
                content={
                    <div className="overflow-x-auto w-full">
                        {backups.length === 0 ? (
                            <p className="text-gray-400 py-4">No backups found. Create a backup to get started.</p>
                        ) : (
                            <table className="w-full">
                                <thead>
                                    <tr className="text-left py-1">
                                        <th className="pr-4">Filename</th>
                                        <th className="pr-4">Original Save</th>
                                        <th className="pr-4">Size</th>
                                        <th className="pr-4">Created</th>
                                        <th className="pr-4">Type</th>
                                        <th className="pr-4">Status</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {backups.map(backup => (
                                        <tr className="py-2 md:py-1" key={backup.id}>
                                            <td className="pr-4">{backup.filename}</td>
                                            <td className="pr-4">{backup.original_save || '-'}</td>
                                            <td className="pr-4">{formatFileSize(backup.size || 0)}</td>
                                            <td className="pr-4">
                                                {backup.created_at
                                                    ? new Date(backup.created_at).toLocaleString()
                                                    : '-'}
                                            </td>
                                            <td className="pr-4">{getTypeBadge(backup.type)}</td>
                                            <td className="pr-4">{getStatusBadge(backup.status)}</td>
                                            <td className="whitespace-nowrap">
                                                <FontAwesomeIcon
                                                    className={`mr-2 cursor-pointer ${
                                                        serverStatus.status === 'running' || isRestoring === backup.id
                                                            ? 'text-gray-500 cursor-not-allowed'
                                                            : 'text-blue-400 hover:text-blue-300'
                                                    }`}
                                                    icon={faUndo}
                                                    title="Restore"
                                                    onClick={() => restoreBackup(backup)}
                                                />
                                                <a href={backupsResource.download(backup.id)} className="mr-2">
                                                    <FontAwesomeIcon
                                                        className="text-gray-light cursor-pointer hover:text-orange"
                                                        icon={faDownload}
                                                        title="Download"
                                                    />
                                                </a>
                                                <FontAwesomeIcon
                                                    className={`mr-2 cursor-pointer ${
                                                        isDeleting === backup.id
                                                            ? 'text-gray-500 cursor-not-allowed'
                                                            : 'text-red hover:text-red-light'
                                                    }`}
                                                    icon={faTrashAlt}
                                                    title="Delete"
                                                    onClick={() => deleteBackup(backup)}
                                                />
                                                <FontAwesomeIcon
                                                    className={`cursor-pointer ${
                                                        isVerifying === backup.id
                                                            ? 'text-gray-500 cursor-not-allowed'
                                                            : 'text-green hover:text-green-light'
                                                    }`}
                                                    icon={faCheckCircle}
                                                    title="Verify"
                                                    onClick={() => verifyBackup(backup)}
                                                />
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </div>
                }
                actions={
                    <Button
                        onClick={updateBackupList}
                        size="sm"
                        type="default"
                    >
                        Refresh
                    </Button>
                }
            />
        </>
    );
};

export default Backups;
