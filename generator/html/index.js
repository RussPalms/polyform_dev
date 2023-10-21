let clientID = null;


function fetch(theUrl, callback) {
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function () {
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
            console.log(xmlHttp.responseText)
            callback(JSON.parse(xmlHttp.responseText));
        }
    }
    xmlHttp.open("GET", theUrl, true); // true for asynchronous 
    xmlHttp.send(null);
}


function post(theUrl, body, callback) {
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function () {
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
            callback(JSON.parse(xmlHttp.responseText));
        }
    }
    xmlHttp.open("POST", theUrl, true); // true for asynchronous 
    xmlHttp.send(JSON.stringify(body));
}

let initID = null
setInterval(() => {
    fetch("/started", (payload) => {
        if (initID === null) {
            initID = payload.time;
        }

        if (initID !== payload.time) {
            location.reload();
        }
    })
}, 1000);


// https://threejs.org/examples/?q=Directional#webgl_lights_hemisphere
// https://threejs.org/examples/#webgl_geometry_spline_editor

const container = document.createElement('div');
document.body.appendChild(container);

import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import Stats from 'three/addons/libs/stats.module.js';
import { GUI } from 'three/addons/libs/lil-gui.module.min.js';
import { CSS2DRenderer, CSS2DObject } from 'three/addons/renderers/CSS2DRenderer.js';
// import { RoomEnvironment } from 'three/addons/environments/RoomEnvironment.js';
import { ProgressiveLightMap } from 'three/addons/misc/ProgressiveLightMap.js';

let viewportSettingsChanged = false;
const viewportSettings = {
    renderWireframe: false,
    fog: {
        color: "0xa0a0a0",
        near: 10,
        far: 50,
    },
    background: "0xa0a0a0",
    lighting: "0xffffff",
    ground: "0xcbcbcb"
}

const shadowMapRes = 4098, lightMapRes = 4098, lightCount = 8;

const clock = new THREE.Clock();

const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
camera.position.set(0, 2, 3);

const scene = new THREE.Scene();
// scene.background = new THREE.Color(viewportSettings.background);

const textureLoader = new THREE.TextureLoader();
const textureEquirec = textureLoader.load('https://i.imgur.com/FFkjGWG_d.png?maxwidth=1520&fidelity=grand');
textureEquirec.mapping = THREE.EquirectangularReflectionMapping;
textureEquirec.colorSpace = THREE.SRGBColorSpace;

// scene.background = textureEquirec;
scene.fog = new THREE.Fog(viewportSettings.fog.color, viewportSettings.fog.near, viewportSettings.fog.far);


const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setPixelRatio(window.devicePixelRatio);
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.shadowMap.enabled = true;
renderer.shadowMap.type = THREE.PCFSoftShadowMap; // default THREE.PCFShadowMap
renderer.toneMapping = THREE.ACESFilmicToneMapping;
renderer.toneMappingExposure = 1;

container.appendChild(renderer.domElement);
// progressive lightmap
const progressiveSurfacemap = new ProgressiveLightMap(renderer, lightMapRes);

const labelRenderer = new CSS2DRenderer();
labelRenderer.setSize(window.innerWidth, window.innerHeight);
labelRenderer.domElement.style.position = 'absolute';
labelRenderer.domElement.style.top = '0px';
labelRenderer.domElement.style.pointerEvents = 'none';
container.appendChild(labelRenderer.domElement);

const hemiLight = new THREE.HemisphereLight(viewportSettings.lighting, 0x8d8d8d, 3);
hemiLight.position.set(0, 20, 0);
scene.add(hemiLight);

const dirLight = new THREE.DirectionalLight(viewportSettings.lighting, 3);
dirLight.position.set(100, 100, 100);
dirLight.castShadow = true;
dirLight.shadow.camera.top = 100;
dirLight.shadow.camera.bottom = -100;
dirLight.shadow.camera.left = - 100;
dirLight.shadow.camera.right = 100;
// dirLight.shadow.camera.far = 40;
dirLight.shadow.camera.near = 0.1;
dirLight.shadow.mapSize.width = shadowMapRes;
dirLight.shadow.mapSize.height = shadowMapRes;
progressiveSurfacemap.addObjectsToLightMap([dirLight])
scene.add(dirLight);


// ground
const groundMat = new THREE.MeshPhongMaterial({ color: viewportSettings.ground, depthWrite: true });
const groundMesh = new THREE.Mesh(new THREE.PlaneGeometry(100, 100), groundMat);
groundMesh.rotation.x = - Math.PI / 2;
groundMesh.receiveShadow = true;
scene.add(groundMesh);
progressiveSurfacemap.addObjectsToLightMap([groundMesh])

const panel = new GUI({ width: 310 });

const updateProfile = (cb) => {
    return post("/profile", profile, () => {
        RefreshProducerOutput();
        if (cb) {
            cb();
        }
    })
}

const allMeshGUISettings = [];
const parseGroupParameters = (folderToAddTo, groupParameters) => {
    const folderSettings = {};

    groupParameters.parameters.forEach((param) => {
        switch (param.type.toLowerCase()) {
            case "group":
                const subFold = parseGroupParameters(folderToAddTo.addFolder(param.name), param);
                if (Object.keys(subFold).length != 0) {
                    folderSettings[param.name] = subFold;
                }
                break;

            case "float":
                folderSettings[param.name] = param.currentValue;
                break;

            case "int":
                folderSettings[param.name] = param.currentValue;
                break;

            case "color":
                folderSettings[param.name] = param.currentValue;
                break;

            case "bool":
                folderSettings[param.name] = param.currentValue;
                break;

            case "string":
                folderSettings[param.name] = param.currentValue;
                break;

            default:
                console.warn("unrecognized param type", param.type, "ignoring", param.name)
        }
    })

    groupParameters.parameters.forEach((param) => {
        let setting = null;
        switch (param.type.toLowerCase()) {
            case "float":
                setting = folderToAddTo.add(folderSettings, param.name).listen().onChange((weight) => {
                    updateProfile();
                });
                break;

            case "int":
                setting = folderToAddTo.add(folderSettings, param.name).step(1).listen().onChange((weight) => {
                    updateProfile();
                });
                break;

            case "string":
                setting = folderToAddTo.add(folderSettings, param.name).step(1).listen().onChange((weight) => {
                    updateProfile();
                });
                break;

            case "bool":
                setting = folderToAddTo.add(folderSettings, param.name).step(1).listen().onChange((weight) => {
                    updateProfile();
                });
                break;

            case "color":
                setting = folderToAddTo.addColor(folderSettings, param.name).listen().onChange((weight) => {
                    updateProfile();
                });
                break;
        }
        if (setting != null) {
            allMeshGUISettings.push(setting);
        }
    });

    return folderSettings;
}

const parseSchemaParameters = (folderToAddTo, curSchema) => {

    const folderSettings = {};

    const subFold = parseGroupParameters(folderToAddTo, curSchema.parameters);
    if (Object.keys(subFold).length != 0) {
        folderSettings["Parameters"] = subFold;
    }

    folderSettings["subGenerators"] = {}

    for (let key of Object.keys(curSchema.subGenerators)) {
        const subFolderToAddTo = folderToAddTo.addFolder(key)
        const subFoldData = parseSchemaParameters(subFolderToAddTo, curSchema.subGenerators[key]);
        if (Object.keys(subFold).length != 0) {
            folderSettings["subGenerators"][key] = subFoldData;
        }
    }

    if (Object.keys(folderSettings).length == 0) {
        return;
    }

    return folderSettings;
}

let schema = {}

let firstTimeLoadingScene = true;

const loader = new GLTFLoader().setPath('producer/');
let producerScene = null;
const RefreshProducerOutput = () => {

    if (producerScene != null) {
        scene.remove(producerScene)
    }

    producerScene = new THREE.Group();
    scene.add(producerScene);
    schema.producers.forEach(producer => {
        const fileExt = producer.split('.').pop().toLowerCase();

        switch (fileExt) {
            case "gltf":
            case "glb":
                loader.load(producer, function (gltf) {
                    producerScene.add(gltf.scene);

                    const aabb = new THREE.Box3();
                    aabb.setFromObject(gltf.scene);
                    const aabbDepth = (aabb.max.z - aabb.min.z)
                    const aabbWidth = (aabb.max.x - aabb.min.x)
                    const aabbHeight = (aabb.max.y - aabb.min.y)
                    const aabbHalfHeight = aabbHeight / 2
                    const mid = (aabb.max.y + aabb.min.y) / 2

                    // We have to do this weird thing because the pivot of the scene
                    // Isn't always the center of the AABB
                    gltf.scene.position.set(0, - mid + aabbHalfHeight, 0)

                    const objects = [];

                    gltf.scene.traverse(function (object) {
                        if (object.isMesh) {
                            object.castShadow = true;
                            object.receiveShadow = true;

                            const prevMaterial = object.material;

                            // if (object.material.userData && object.material.userData["threejs-material"] === "phong") {
                            //     object.material = new THREE.MeshPhongMaterial();

                            // } else {
                            // object.material = new THREE.MeshPhysicalMaterial();
                            // }

                            // THREE.MeshBasicMaterial.prototype.copy.call( object.material, prevMaterial );

                            // // Copying what I want...
                            // object.material.color = prevMaterial.color;
                            // object.materialroughness = prevMaterial.roughness;
                            // object.materialroughnessMap = prevMaterial.roughnessMap;
                            // object.materialmetalness = prevMaterial.metalness;
                            // object.materialmetalnessMap = prevMaterial.metalnessMap;

                            object.material.wireframe = viewportSettings.renderWireframe;
                            object.material.envMap = textureEquirec;
                            object.material.needsUpdate = true;
                            object.material.transparent = true;

                            console.log(prevMaterial)
                            objects.push(object)
                        }
                    });

                    progressiveSurfacemap.addObjectsToLightMap(objects);

                    if (firstTimeLoadingScene) {
                        firstTimeLoadingScene = false;

                        camera.position.y = mid * (3 / 2);
                        camera.position.z = Math.sqrt(
                            (aabbWidth * aabbWidth) +
                            (aabbDepth * aabbDepth) +
                            (aabbHeight * aabbHeight)
                        ) / 2;

                        controls.target.set(0, mid, 0);
                        controls.update();
                    }

                });
                break;
        }



    });
}

const fileControls = {
    saveProfile: () => {
        const fileContent = JSON.stringify(profile);
        const bb = new Blob([fileContent], { type: 'application/json' });
        const a = document.createElement('a');
        a.download = 'profile.json';
        a.href = window.URL.createObjectURL(bb);
        a.click();
    },
    loadProfile: () => {
        const input = document.createElement('input');
        input.type = 'file';

        input.onchange = e => {

            // getting a hold of the file reference
            const file = e.target.files[0];

            // setting up the reader
            const reader = new FileReader();
            reader.readAsText(file, 'UTF-8');

            // here we tell the reader what to do when it's done reading...
            reader.onload = readerEvent => {
                const content = readerEvent.target.result; // this is the content!
                console.log(content);
                profile = JSON.parse(content)
                updateProfile(() => {
                    location.reload();
                })
            }

        }

        input.click();
    }
}
const fileSettingsFolder = panel.addFolder("File");
fileSettingsFolder.add(fileControls, "saveProfile").name("Save Profile")
fileSettingsFolder.add(fileControls, "loadProfile").name("Load Profile")


const viewportSettingsFolder = panel.addFolder("Render Settings");
const viewportManager = {}

const wireFrameUpdater = () => {
    if (producerScene == null) {
        return;
    }
    producerScene.traverse((object) => {
        if (object.isMesh) {
            object.material.wireframe = viewportSettings.renderWireframe;
        }
    });
}
viewportManager["renderWireframe"] = {
    setting: viewportSettingsFolder
        .add(viewportSettings, "renderWireframe")
        .name("Render Wireframe")
        .listen()
        .onChange((weight) => {
            viewportSettingsChanged = true;
            wireFrameUpdater();
        }),
    updater: wireFrameUpdater
}

const backgroundUpdater = () => {
    scene.background = new THREE.Color(viewportSettings.background);
}

viewportManager["background"] = {
    setting: viewportSettingsFolder
        .addColor(viewportSettings, "background")
        .name("Background")
        .listen()
        .onChange((weight) => {
            viewportSettingsChanged = true;
            backgroundUpdater();
        }),
    updater: backgroundUpdater
}


const lightingUpdater = () => {
    dirLight.color = new THREE.Color(viewportSettings.lighting);
    hemiLight.color = new THREE.Color(viewportSettings.lighting);
}

viewportManager["lighting"] = {
    setting: viewportSettingsFolder
        .addColor(viewportSettings, "lighting")
        .name("Lighting")
        .listen()
        .onChange((weight) => {
            viewportSettingsChanged = true;
            lightingUpdater();
        }),
    updater: lightingUpdater
}

const groundUpdater = () => {
    groundMat.color = new THREE.Color(viewportSettings.ground);
}

viewportManager["ground"] = {
    setting: viewportSettingsFolder
        .addColor(viewportSettings, "ground")
        .name("Ground")
        .listen()
        .onChange((weight) => {
            viewportSettingsChanged = true;
            groundUpdater();
        }),
    updater: groundUpdater
}

const fogSettingsManager = {}

const fogSettingsFolder = viewportSettingsFolder.addFolder("Fog");
fogSettingsFolder.close();


const fogColorUpdater = () => {
    scene.fog.color = new THREE.Color(viewportSettings.fog.color);
}
fogSettingsManager["color"] = {
    setting: fogSettingsFolder.addColor(viewportSettings.fog, "color")
        .listen()
        .onChange((_) => {
            viewportSettingsChanged = true;
            fogColorUpdater();
        }),
    updater: fogColorUpdater
}


const fogNearUpdater = () => {
    scene.fog.near = viewportSettings.fog.near;
}
fogSettingsManager["near"] = {
    setting: fogSettingsFolder.add(viewportSettings.fog, "near")
        .listen()
        .onChange((_) => {
            viewportSettingsChanged = true;
            fogNearUpdater();
        }),
    updater: fogNearUpdater
}


const fogFarUpdater = () => {
    scene.fog.far = viewportSettings.fog.far;
}
fogSettingsManager["far"] = {
    setting: fogSettingsFolder.add(viewportSettings.fog, "far")
        .listen()
        .onChange((_) => {
            viewportSettingsChanged = true;
            fogFarUpdater();
        }),
    updater: fogFarUpdater
}


let profile = {};

fetch("/schema", (generatorSchema) => {
    schema = generatorSchema;
    RefreshProducerOutput();
    profile = parseSchemaParameters(panel.addFolder("Mesh Generation"), generatorSchema)
})

const updateProfileParametersWithNewSchema = (prof, newSchema) => {
    newSchema.parameters.forEach(schemaParam => {

        switch (schemaParam.type) {
            case "Group":
                updateProfileParametersWithNewSchema(prof[schemaParam.name], schemaParam)
                break;

            default:
                prof[schemaParam.name] = schemaParam.currentValue;
                break;
        }
    });
}

const updateProfileWithNewSchema = (prof, newSchema) => {
    for (const [key, gen] of Object.entries(prof.subGenerators)) {
        updateProfileWithNewSchema(gen, newSchema.subGenerators[key])
    }
    console.log(prof)
    updateProfileParametersWithNewSchema(prof.Parameters, newSchema.parameters);
}

const featchandApplyLatestSchemaToControls = () => {
    fetch("/schema", (generatorSchema) => {
        schema = generatorSchema;
        RefreshProducerOutput();
        updateProfileWithNewSchema(profile, generatorSchema)
        allMeshGUISettings.forEach(setting => {
            setting.updateDisplay();
        });
    })
}


// const environment = new RoomEnvironment(renderer);
// const pmremGenerator = new THREE.PMREMGenerator(renderer);
// scene.environment = pmremGenerator.fromScene( environment ).texture;

const controls = new OrbitControls(camera, renderer.domElement);
// controls.addEventListener('change', render); // use if there is no animation loop
controls.minDistance = 0;
controls.maxDistance = 100;
controls.target.set(0, 0, 0);
controls.update();

camera.position.z = 5;

const stats = new Stats();
container.appendChild(stats.dom);


function onWindowResize() {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
    labelRenderer.setSize(window.innerWidth, window.innerHeight);
}

window.addEventListener('resize', onWindowResize);

const connectedPlayers = {}

const playerGeometry = new THREE.SphereGeometry(1, 32, 16);
const playerMaterial = new THREE.MeshPhongMaterial({ color: 0xffff00 });
const playerEyeMaterial = new THREE.MeshBasicMaterial({ color: 0x000000 });

let lastUpdatedModel = -1;
if (window["WebSocket"]) {
    const conn = new WebSocket("ws://" + document.location.host + "/live");
    conn.onclose = function (evt) {
        console.log("connection closed", evt)
    };
    conn.onmessage = function (evt) {
        const message = JSON.parse(evt.data);
        // console.log("on message", message)

        switch (message.type) {
            case "Server-SetClientID":
                clientID = message.data;
                break;

            case "Server-RoomStateUpdate":
                if (lastUpdatedModel !== message.data.ModelVersion) {
                    lastUpdatedModel = message.data.ModelVersion;
                    featchandApplyLatestSchemaToControls();
                }

                if (viewportSettingsChanged === false) {
                    const webScene = message.data.WebScene;
                    // console.log(webScene)

                    for (const [setting, data] of Object.entries(viewportManager)) {
                        if (viewportSettings[setting] !== webScene[setting]) {
                            viewportSettings[setting] = webScene[setting];
                            console.log(setting)
                            data.setting.updateDisplay();
                            data.updater();
                        }
                    }

                    for (const [setting, data] of Object.entries(fogSettingsManager)) {
                        if (viewportSettings.fog[setting] !== webScene.fog[setting]) {
                            viewportSettings.fog[setting] = webScene.fog[setting];
                            console.log(setting)
                            data.setting.updateDisplay();
                            data.updater();
                        }
                    }
                }

                const playersUpdated = {}
                for (const [key, value] of Object.entries(connectedPlayers)) {
                    playersUpdated[key] = false;
                }

                for (const [key, value] of Object.entries(message.data.Players)) {
                    if (value == null) {
                        continue;
                    }

                    // We don't want to create a representation of ourselves
                    if (key == clientID) {
                        continue;
                    }

                    playersUpdated[key] = true;

                    if (key in connectedPlayers) {
                        // Update the player we've already instantiated
                        connectedPlayers[key].desiredPosition.x = value.position.x;
                        connectedPlayers[key].desiredPosition.y = value.position.y;
                        connectedPlayers[key].desiredPosition.z = value.position.z;

                        connectedPlayers[key].desiredRotation.x = value.rotation.x;
                        connectedPlayers[key].desiredRotation.y = value.rotation.y;
                        connectedPlayers[key].desiredRotation.z = value.rotation.z;
                        connectedPlayers[key].desiredRotation.w = value.rotation.w;
                    } else {
                        // Create a new Player!
                        const newPlayer = new THREE.Group();

                        const sphere = new THREE.Mesh(playerGeometry, playerMaterial);
                        sphere.position.z += 0.5;
                        newPlayer.add(sphere);


                        const eyeSize = 0.15;
                        const eyeSpacing = 0.3;

                        const leftEye = new THREE.Mesh(playerGeometry, playerEyeMaterial);
                        leftEye.scale.x = eyeSize;
                        leftEye.scale.y = eyeSize;
                        leftEye.scale.z = eyeSize;
                        leftEye.position.x = eyeSpacing;
                        leftEye.position.z = - 0.5;
                        leftEye.position.y = + 0.25;
                        newPlayer.add(leftEye);

                        const rightEye = new THREE.Mesh(playerGeometry, playerEyeMaterial);
                        rightEye.scale.x = eyeSize;
                        rightEye.scale.y = eyeSize;
                        rightEye.scale.z = eyeSize;
                        rightEye.position.x = - eyeSpacing;
                        rightEye.position.z = - 0.5;
                        rightEye.position.y = + 0.25;
                        newPlayer.add(rightEye);

                        const text = document.createElement('div');
                        text.className = 'label';
                        text.style.color = '#000000';
                        text.textContent = value.name;
                        text.style.fontSize = "30px";

                        const label = new CSS2DObject(text);
                        label.position.y += 0.75;
                        newPlayer.add(label);

                        connectedPlayers[key] = {
                            obj: newPlayer,
                            desiredPosition: value.position,
                            desiredRotation: value.rotation,
                            label: label
                        };

                        newPlayer.position.x = value.position.x;
                        newPlayer.position.y = value.position.y;
                        newPlayer.position.z = value.position.z;
                        scene.add(newPlayer);
                    }
                }

                // Remove all players that weren't contained within the update
                for (const [playerID, updated] of Object.entries(playersUpdated)) {
                    if (updated) {
                        continue;
                    }

                    // We need to explicitly call remove on the label 
                    // so it cleans up the DOM
                    connectedPlayers[playerID].obj.remove(connectedPlayers[playerID].label);

                    scene.remove(connectedPlayers[playerID].obj);
                    delete connectedPlayers[playerID];
                }

                break;

            case "Server-RefreshGenerator":
                break;

            case "Server-Broadcast":
                break;
        }

    };

    setInterval(() => {
        // console.log(camera.position)
        conn.send(JSON.stringify({
            "type": "Client-SetOrientation",
            "data": {
                "position": {
                    "x": camera.position.x,
                    "y": camera.position.y,
                    "z": camera.position.z,
                },
                "rotation": {
                    "x": camera.rotation.x,
                    "y": camera.rotation.y,
                    "z": camera.rotation.z,
                    "w": camera.rotation.w,
                }
            }
        }));
    }, 200);

    setInterval(() => {
        if (viewportSettingsChanged === false) {
            return;
        }
        console.log("updating...")
        viewportSettingsChanged = false;
        console.log(viewportSettings)
        conn.send(JSON.stringify({
            "type": "Client-SetScene",
            "data": viewportSettings
        }));
    }, 200);
} else {
    console.error("web browser does not support web sockets")
}


function animate() {
    const delta = clock.getDelta();

    for (const [key, player] of Object.entries(connectedPlayers)) {
        const pr = player.obj.rotation;
        const dr = player.desiredRotation;

        pr.x = pr.x + ((dr.x - pr.x) * delta * 2)
        pr.y = pr.y + ((dr.y - pr.y) * delta * 2)
        pr.z = pr.z + ((dr.z - pr.z) * delta * 2)
        pr.w = pr.w + ((dr.w - pr.w) * delta * 2)


        const pp = player.obj.position;
        const dp = player.desiredPosition;

        pp.x = pp.x + ((dp.x - pp.x) * delta * 2)
        pp.y = pp.y + ((dp.y - pp.y) * delta * 2)
        pp.z = pp.z + ((dp.z - pp.z) * delta * 2)
    }

    requestAnimationFrame(animate);
    renderer.render(scene, camera);
    labelRenderer.render(scene, camera);

    stats.update();
}
animate();