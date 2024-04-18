const panel = new GUI({ width: 310 });

import * as GaussianSplats3D from '@mkkellogg/gaussian-splats-3d';


let initID = null
setInterval(() => {
    requestManager.getStartedTime((payload) => {
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

const container = document.getElementById('three-viewer-container');

import * as THREE from 'three';
import { NodeManager } from "./node_manager.js";
import { WebSocketManager, WebSocketRepresentationManager } from "./websocket.js";
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import Stats from 'three/addons/libs/stats.module.js';
import { GUI } from 'three/addons/libs/lil-gui.module.min.js';
import { CSS2DRenderer } from 'three/addons/renderers/CSS2DRenderer.js';
// import { RoomEnvironment } from 'three/addons/environments/RoomEnvironment.js';
import { ProgressiveLightMap } from 'three/addons/misc/ProgressiveLightMap.js';

import { InitXR } from './xr.js';
import { UpdateManager } from './update-manager.js';
import { ColorSelector } from './color_selector.js';

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

const representationManager = new WebSocketRepresentationManager();
const viewportManager = new ViewportManager(viewportSettings);
const updateLoop = new UpdateManager();

const shadowMapRes = 4098, lightMapRes = 4098, lightCount = 8;

const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
camera.position.set(0, 2, 3);

representationManager.AddRepresentation(0, camera)

const scene = new THREE.Scene();
scene.background = new THREE.Color(viewportSettings.background);

const textureLoader = new THREE.TextureLoader();
const textureEquirec = textureLoader.load('https://i.imgur.com/FFkjGWG_d.png?maxwidth=1520&fidelity=grand');
textureEquirec.mapping = THREE.EquirectangularReflectionMapping;
textureEquirec.colorSpace = THREE.SRGBColorSpace;

// scene.background = textureEquirec;
scene.fog = new THREE.Fog(viewportSettings.fog.color, viewportSettings.fog.near, viewportSettings.fog.far);

const viewerContainer = new THREE.Group();
scene.add(viewerContainer);

const threeCanvas = document.getElementById("three-canvas");
const renderer = new THREE.WebGLRenderer({ 
    canvas: threeCanvas,
    antialias: RenderingConfiguration.AntiAlias 
});
renderer.setPixelRatio(window.devicePixelRatio);
renderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
renderer.shadowMap.enabled = true;
renderer.shadowMap.type = THREE.PCFSoftShadowMap; // default THREE.PCFShadowMap
renderer.toneMapping = THREE.ACESFilmicToneMapping;
renderer.toneMappingExposure = 1;
renderer.xr.enabled = RenderingConfiguration.XrEnabled;
renderer.setAnimationLoop(updateLoop.run.bind(updateLoop))

// container.appendChild(renderer.domElement);
// progressive lightmap
const progressiveSurfacemap = new ProgressiveLightMap(renderer, lightMapRes);

const labelRenderer = new CSS2DRenderer();
labelRenderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
labelRenderer.domElement.style.position = 'absolute';
labelRenderer.domElement.style.top = '0px';
labelRenderer.domElement.style.pointerEvents = 'none';
container.appendChild(labelRenderer.domElement);

const stats = new Stats();
container.appendChild(stats.dom);

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

// const environment = new RoomEnvironment(renderer);
// const pmremGenerator = new THREE.PMREMGenerator(renderer);
// scene.environment = pmremGenerator.fromScene( environment ).texture;

const orbitControls = new OrbitControls(camera, renderer.domElement);
// controls.addEventListener('change', render); // use if there is no animation loop
orbitControls.minDistance = 0;
orbitControls.maxDistance = 100;
orbitControls.target.set(0, 0, 0);
orbitControls.update();

camera.position.z = 5;

const requestManager = new RequestManager();

const App = {
    Camera: camera,
    Renderer: renderer,
    // MeshGenFolder: panel.addFolder("Mesh Generation"),
    Scene: scene,
    OrbitControls: orbitControls,
    ViewerScene: viewerContainer,
    LightGraph: lgraphInstance,
    ColorSelector: new ColorSelector("colorSelectorContainer"),
    RequestManager: requestManager,
    ServerUpdatingNodeConnections: false,
}

const nodeManager = new NodeManager(App);
const schemaManager = new SchemaManager(requestManager, nodeManager);

if (RenderingConfiguration.XrEnabled) {
    InitXR(scene, renderer, updateLoop, representationManager, groundMesh);
}

nodeManager.subscribeToParameterChange((param) => {
    console.log(param)
    schemaManager.setProfileKey(param.id, param.data, param.binary);
});

let firstTimeLoadingScene = true;

const loader = new GLTFLoader().setPath('producer/');
let producerScene = null;

let guassianSplatViewer = null;


class SchemaRefreshManager {
    constructor() {
        this.loadingCount = 0;
        this.cachedSchema = null;
    }

    AddLoading() {
        console.log("add")
        this.loadingCount += 1;
    }

    RemoveLoading() {
        console.log("remove")
        if (this.loadingCount === 0) {
            throw new Error("loading count already 0");
        }
        this.loadingCount -= 1;

        if (this.loadingCount === 0 && this.cachedSchema) {
            this.Refresh(this.cachedSchema)
            this.cachedSchema = null;
        }
    }

    CurrentlyLoading() {
        return this.loadingCount > 0;
    }

    NewSchema(schema) {
        if (this.CurrentlyLoading()) {
            this.cachedSchema = schema;
            return;
        }
        this.Refresh(schema);
    }

    loadText(producerURL) {
        this.AddLoading();
        requestManager.fetchText(
            producerURL,
            (data) => {
                InfoManager.ShowInfo(data);
                this.RemoveLoading();
            },
            (error) => {
                this.RemoveLoading();
            }
        );
    }

    loadGltf(producerURL) {
        this.AddLoading();
        loader.load(producerURL, (gltf) => {

            const aabb = new THREE.Box3();
            aabb.setFromObject(gltf.scene);
            const aabbDepth = (aabb.max.z - aabb.min.z)
            const aabbWidth = (aabb.max.x - aabb.min.x)
            const aabbHeight = (aabb.max.y - aabb.min.y)
            const aabbHalfHeight = aabbHeight / 2
            const mid = (aabb.max.y + aabb.min.y) / 2

            producerScene.add(gltf.scene);

            // We have to do this weird thing because the pivot of the scene
            // Isn't always the center of the AABB
            viewerContainer.position.set(0, - mid + aabbHalfHeight, 0)

            const objects = [];

            gltf.scene.traverse((object) => {
                console.log(object)
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
                } else if (object.isPoints) {
                    object.material.size = 2;
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

                orbitControls.target.set(0, mid, 0);
                orbitControls.update();
            }
            this.RemoveLoading();
        },
            undefined,
            (error) => {
                this.RemoveLoading();
                error.response.json().then(x => {
                    ErrorManager.ShowError(x.error);
                })
            });
    }

    loadSplat(producerURL) {
        this.AddLoading();
        if (guassianSplatViewer) {
            guassianSplatViewer.dispose();
        }

        renderer.setPixelRatio(1);

        // https://github.com/mkkellogg/GaussianSplats3D/blob/main/src/Viewer.js
        const splatViewerOptions = {
            selfDrivenMode: false,
            useBuiltInControls: false,
            rootElement: renderer.domElement.parentElement,
            renderer: renderer,
            threeScene: scene,
            camera: camera,
            gpuAcceleratedSort: true,
            // 'sceneRevealMode': GaussianSplats3D.SceneRevealMode.Instant,
            sharedMemoryForWebWorkers: true
        }

        guassianSplatViewer = new GaussianSplats3D.Viewer(splatViewerOptions);

        // getSplatCenter
        guassianSplatViewer.addSplatScene(producerURL, {
            // streamView: false
            // 'scale': [0.25, 0.25, 0.25],
        }).then(() => {

            guassianSplatViewer.splatMesh.onSplatTreeReady((splatTree) => {
                const tree = splatTree.subTrees[0]
                const aabb = new THREE.Box3();
                aabb.setFromPoints([tree.sceneMin, tree.sceneMax]);
                const aabbDepth = (aabb.max.z - aabb.min.z)
                const aabbWidth = (aabb.max.x - aabb.min.x)
                const aabbHeight = (aabb.max.y - aabb.min.y)
                const aabbHalfHeight = aabbHeight / 2
                const mid = (aabb.max.y + aabb.min.y) / 2
    
                const shiftY = - mid + aabbHalfHeight
                guassianSplatViewer.splatMesh.position.set(0, shiftY, 0)
                viewerContainer.position.set(0, shiftY, 0)
    
                if (firstTimeLoadingScene) {
                    firstTimeLoadingScene = false;
    
                    camera.position.y = mid * (3 / 2);
                    camera.position.z = Math.sqrt(
                        (aabbWidth * aabbWidth) +
                        (aabbDepth * aabbDepth) +
                        (aabbHeight * aabbHeight)
                    ) / 2;
    
                    orbitControls.target.set(0, mid, 0);
                    orbitControls.update();
                }
           });

            this.RemoveLoading();
        }).catch(x => {
            console.error(x)
            this.RemoveLoading();
        })
    }

    Refresh(schema) {
        ErrorManager.ClearError();
        InfoManager.ClearInfo();

        if (producerScene != null) {
            viewerContainer.remove(producerScene)
        }

        producerScene = new THREE.Group();
        viewerContainer.add(producerScene);

        for (const [producer, producerData] of Object.entries(schema.producers)) {
            const fileExt = producer.split('.').pop().toLowerCase();

            switch (fileExt) {
                case "txt":
                    this.loadText('producer/' + producer);
                    break;

                case "gltf":
                case "glb":
                    this.loadGltf(producer);
                    break;

                case "splat":
                    this.loadSplat('producer/' + producer)
                    break;
            }
        }
    }
}

const schemaRefreshManager = new SchemaRefreshManager();

schemaManager.subscribe(schemaRefreshManager.NewSchema.bind(schemaRefreshManager));



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
                profile = JSON.parse(content)
                updateProfile(() => {
                    location.reload();
                })
            }

        }

        input.click();
    },
    saveModel: () => {
        download("/zip", (data) => {
            const a = document.createElement('a');
            a.download = 'model.zip';
            const url = window.URL.createObjectURL(data);
            a.href = url;
            a.click();
            window.URL.revokeObjectURL(url);
        })
    },
    viewProgram: () => {
        requestManager.fetchText("/mermaid", (data) => {
            const mermaidConfig = {
                "code": data,
                "mermaid": {
                    "securityLevel": "strict"
                }
            }

            const mermaidURL = "https://mermaid.live/edit#" + btoa(JSON.stringify(mermaidConfig));
            window.open(mermaidURL, '_blank').focus();
        })
    }
}

const fileSettingsFolder = panel.addFolder("File");
fileSettingsFolder.add(fileControls, "saveProfile").name("Save Profile")
fileSettingsFolder.add(fileControls, "loadProfile").name("Load Profile")
fileSettingsFolder.add(fileControls, "saveModel").name("Download Model")
fileSettingsFolder.add(fileControls, "viewProgram").name("View Program")
fileSettingsFolder.close();

const viewportSettingsFolder = panel.addFolder("Rendering");
viewportSettingsFolder.close();

viewportManager.AddSetting(
    "renderWireframe",
    new ViewportSetting(
        "renderWireframe",
        viewportSettings,
        viewportSettingsFolder
            .add(viewportSettings, "renderWireframe")
            .name("Render Wireframe"),
        () => {
            if (producerScene == null) {
                return;
            }
            producerScene.traverse((object) => {
                if (object.isMesh) {
                    object.material.wireframe = viewportSettings.renderWireframe;
                }
            });
        }
    )
)


viewportManager.AddSetting(
    "background",
    new ViewportSetting(
        "background",
        viewportSettings,
        viewportSettingsFolder
            .addColor(viewportSettings, "background")
            .name("Background"),
        () => {
            scene.background = new THREE.Color(viewportSettings.background);
        }
    )
);


viewportManager.AddSetting(
    "lighting",
    new ViewportSetting(
        "lighting",
        viewportSettings,
        viewportSettingsFolder
            .addColor(viewportSettings, "lighting")
            .name("Lighting"),
        () => {
            dirLight.color = new THREE.Color(viewportSettings.lighting);
            hemiLight.color = new THREE.Color(viewportSettings.lighting);
        },
    )
);

viewportManager.AddSetting(
    "ground",
    new ViewportSetting(
        "ground",
        viewportSettings,
        viewportSettingsFolder
            .addColor(viewportSettings, "ground")
            .name("Ground"),
        () => {
            groundMat.color = new THREE.Color(viewportSettings.ground);
        }
    )
);


const fogSettingsFolder = viewportSettingsFolder.addFolder("Fog");
fogSettingsFolder.close();

viewportManager.AddSetting(
    "fog/color",
    new ViewportSetting(
        "color",
        viewportSettings.fog,
        fogSettingsFolder.addColor(viewportSettings.fog, "color"),
        () => {
            scene.fog.color = new THREE.Color(viewportSettings.fog.color);
        }
    )
);

viewportManager.AddSetting(
    "fog/near",
    new ViewportSetting(
        "near",
        viewportSettings.fog,
        fogSettingsFolder.add(viewportSettings.fog, "near"),
        () => {
            scene.fog.near = viewportSettings.fog.near;
        }
    )
);

viewportManager.AddSetting(
    "fog/far",
    new ViewportSetting(
        "far",
        viewportSettings.fog,
        fogSettingsFolder.add(viewportSettings.fog, "far"),
        () => {
            scene.fog.far = viewportSettings.fog.far;
        }
    )
);


function resize() {
    const w = renderer.domElement.clientWidth;
    const h = renderer.domElement.clientHeight
    
    if (renderer.domElement.width !== w || renderer.domElement.height !== h) {
        renderer.setSize(w, h, false);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
        lightCanvas.resize(lightCanvasCanvas.clientWidth, lightCanvasCanvas.clientHeight, false)
        labelRenderer.setSize(w, h);
    }
}

const websocketManager = new WebSocketManager(
    representationManager,
    scene,
    {
        playerGeometry: new THREE.SphereGeometry(1, 32, 16),
        playerMaterial: new THREE.MeshPhongMaterial({ color: 0xffff00 }),
        playerEyeMaterial: new THREE.MeshBasicMaterial({ color: 0x000000 }),
    },
    viewportManager,
    schemaManager
);
if (websocketManager.canConnect()) {
    websocketManager.connect();
    updateLoop.addToUpdate(websocketManager.update.bind(websocketManager));
} else {
    console.error("web browser does not support web sockets")
}

updateLoop.addToUpdate(() => {
    resize();
    
    renderer.render(scene, camera);

    if (guassianSplatViewer) {
        guassianSplatViewer.update();
        guassianSplatViewer.render();
    }

    labelRenderer.render(scene, camera);
    stats.update();
});
