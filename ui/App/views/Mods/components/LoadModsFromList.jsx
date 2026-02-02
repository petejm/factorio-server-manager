import React, {useEffect, useState} from "react";
import Button from "../../../components/Button";
import Label from "../../../components/Label";
import {useForm} from "react-hook-form";
import modResource from "../../../../api/resources/mods";
import FactorioLogin from "./AddMod/components/FactorioLogin";
import ConfirmDialog from "../../../components/ConfirmDialog";

const LoadModsFromList = ({refreshMods}) => {

    const defaultFileName = 'Select mod-list.json ...';
    const [fileName, setFileName] = useState(defaultFileName);
    const {register, handleSubmit, reset} = useForm();
    const [isLoading, setIsLoading] = useState(false);
    const [isFactorioAuthenticated, setIsFactorioAuthenticated] = useState(false);
    const [modListData, setModListData] = useState(undefined);
    const [parsedMods, setParsedMods] = useState([]);

    useEffect(() => {
        (async () => {
            setIsFactorioAuthenticated(await modResource.portal.status())
        })();
    }, []);

    const onFileSelect = async (e) => {
        const file = e.currentTarget.files[0];
        if (!file) return;

        setFileName(file.name);

        try {
            const text = await file.text();
            const json = JSON.parse(text);
            if (json.mods && Array.isArray(json.mods)) {
                setParsedMods(json.mods);
            } else {
                window.flash("Invalid mod-list.json format", "red");
                setParsedMods([]);
            }
        } catch (err) {
            window.flash("Error parsing mod-list.json: " + err.message, "red");
            setParsedMods([]);
        }
    };

    const loadModsRequested = () => {
        if (parsedMods.length === 0) {
            window.flash("No mods found in the file", "red");
            return;
        }
        setIsLoading(true);
        setModListData(parsedMods);
    };

    const loadMods = async () => {
        await modResource.deleteAll();

        await modResource.portal.installByName(parsedMods)
            .then(() => {
                refreshMods();
                const enabledCount = parsedMods.filter(m => m.enabled).length;
                window.flash(`Loaded ${enabledCount} mods from mod-list.json`, "green");
            })
            .catch(err => {
                window.flash("Error loading mods: " + err.message, "red");
            })
            .finally(() => {
                setIsLoading(false);
                setModListData(undefined);
                setFileName(defaultFileName);
                setParsedMods([]);
                reset();
            });
    };

    const enabledModCount = parsedMods.filter(m => m.enabled && m.name !== "base" && m.name !== "elevated-rails" && m.name !== "quality" && m.name !== "space-age").length;

    return isFactorioAuthenticated
        ? <form onSubmit={handleSubmit(loadModsRequested)}>
            <Label text="Mod List File" htmlFor="mod_list_file"/>
            <div className="relative bg-white shadow text-black h-full w-full mb-4">
                <input
                    {...register('mod_list_file')}
                    className="absolute left-0 top-0 opacity-0 cursor-pointer w-full h-full"
                    onChange={onFileSelect}
                    id="mod_list_file"
                    type="file"
                    accept=".json,application/json"
                />
                <div className="px-2 py-2">{fileName}</div>
            </div>
            {parsedMods.length > 0 && (
                <div className="mb-4 text-sm">
                    Found {enabledModCount} enabled mods (excluding base game mods)
                </div>
            )}
            <Button isSubmit={true} isDisabled={parsedMods.length === 0} isLoading={isLoading}>
                Load Mods
            </Button>
            <ConfirmDialog
                title="Load Mods from List"
                content={`Loading ${enabledModCount} mods from the mod list will remove all currently installed mods.`}
                isOpen={modListData !== undefined}
                close={() => {
                    setIsLoading(false);
                    setModListData(undefined);
                }}
                onSuccess={loadMods}
            />
        </form>
        : <FactorioLogin setIsFactorioAuthenticated={setIsFactorioAuthenticated}/>
};

export default LoadModsFromList;
