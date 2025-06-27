package templates

import (
	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	"maragu.dev/gomponents/html"
)

func HomePage() g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    "3D Model Generator",
		Language: "en",
		Head: []g.Node{
			html.Meta(g.Attr("charset", "UTF-8")),
			html.Meta(g.Attr("name", "viewport"), g.Attr("content", "width=device-width, initial-scale=1.0")),
			g.El("style", g.Text(css)),
			html.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/three@0.177.0/build/three.module.js")),
			html.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/three@0.177.0/examples/js/loaders/GLTFLoader.js")),
			html.Script(g.Attr("src", "https://cdn.jsdelivr.net/npm/three@0.177.0/examples/js/controls/OrbitControls.js")),
		},
		Body: []g.Node{
			html.Div(g.Attr("class", "app-container"),
				html.Div(g.Attr("class", "sidebar"),
					html.H2(g.Text("Previous Models")),
					html.Div(g.Attr("id", "modelList"), g.Attr("class", "model-list"),
						html.P(g.Text("Loading models...")),
					),
				),
				html.Div(g.Attr("class", "main-content"),
					html.Div(g.Attr("class", "container"),
						html.H1(g.Text("3D Model Generator")),
						html.P(g.Text("Upload an image to generate a 3D model using Stability AI")),

						html.Form(g.Attr("id", "uploadForm"), g.Attr("enctype", "multipart/form-data"),
							html.Div(g.Attr("class", "upload-area"), g.Attr("id", "uploadArea"),
								html.P(g.Text("Drag and drop an image here or click to select")),
								html.Input(
									g.Attr("type", "file"),
									g.Attr("id", "imageInput"),
									g.Attr("name", "image"),
									g.Attr("accept", "image/jpeg,image/jpg,image/png"),
									g.Attr("required"),
								),
							),
							html.Button(
								g.Attr("type", "submit"),
								g.Attr("id", "submitBtn"),
								g.Text("Generate 3D Model"),
							),
						),

						html.Div(g.Attr("id", "status"), g.Attr("class", "status hidden")),
						html.Div(g.Attr("id", "result"), g.Attr("class", "result hidden")),

						html.Div(g.Attr("id", "viewerContainer"), g.Attr("class", "viewer-container hidden"),
							html.H2(g.Text("3D Model Viewer")),
							html.Div(g.Attr("id", "modelViewer"), g.Attr("class", "model-viewer")),
							html.Div(g.Attr("class", "viewer-controls"),
								html.P(g.Text("Use mouse to rotate, scroll to zoom")),
							),
						),
					),
				),
			),
			html.Script(g.Raw(javascript)),
		},
	})
}

const css = `
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    margin: 0;
    padding: 0;
    background-color: #f5f5f5;
}

.app-container {
    display: flex;
    height: 100vh;
}

.sidebar {
    width: 300px;
    background: white;
    border-right: 1px solid #ddd;
    padding: 20px;
    overflow-y: auto;
}

.sidebar h2 {
    margin-top: 0;
    color: #333;
}

.model-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.model-item {
    padding: 10px;
    background: #f5f5f5;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s;
}

.model-item:hover {
    background-color: #e0e0e0;
}

.model-item.active {
    background-color: #e3f2fd;
    border: 1px solid #2196F3;
}

.model-item .model-name {
    font-weight: 500;
    color: #333;
}

.model-item .model-date {
    font-size: 12px;
    color: #666;
}

.main-content {
    flex: 1;
    overflow-y: auto;
}

.container {
    max-width: 600px;
    margin: 50px auto;
    padding: 20px;
    background: white;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

h1 {
    text-align: center;
    color: #333;
}

.upload-area {
    border: 2px dashed #ddd;
    border-radius: 8px;
    padding: 40px;
    text-align: center;
    cursor: pointer;
    transition: all 0.3s;
}

.upload-area:hover {
    border-color: #4CAF50;
    background-color: #f9f9f9;
}

.upload-area.dragover {
    border-color: #4CAF50;
    background-color: #e8f5e9;
}

input[type="file"] {
    display: none;
}

button {
    width: 100%;
    padding: 12px;
    margin-top: 20px;
    background-color: #4CAF50;
    color: white;
    border: none;
    border-radius: 5px;
    font-size: 16px;
    cursor: pointer;
    transition: background-color 0.3s;
}

button:hover:not(:disabled) {
    background-color: #45a049;
}

button:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
}

.status {
    margin-top: 20px;
    padding: 15px;
    border-radius: 5px;
    text-align: center;
}

.status.processing {
    background-color: #e3f2fd;
    color: #1976d2;
}

.status.error {
    background-color: #ffebee;
    color: #c62828;
}

.result {
    margin-top: 20px;
    text-align: center;
}

.result a {
    display: inline-block;
    padding: 10px 20px;
    background-color: #2196F3;
    color: white;
    text-decoration: none;
    border-radius: 5px;
    transition: background-color 0.3s;
}

.result a:hover {
    background-color: #1976d2;
}

.hidden {
    display: none;
}

.viewer-container {
    max-width: 800px;
    margin: 30px auto;
    padding: 20px;
    background: white;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.model-viewer {
    width: 100%;
    height: 500px;
    background: #f0f0f0;
    border-radius: 5px;
    position: relative;
}

.viewer-controls {
    text-align: center;
    margin-top: 10px;
    color: #666;
    font-size: 14px;
}
`

const javascript = `
const uploadArea = document.getElementById('uploadArea');
const imageInput = document.getElementById('imageInput');
const uploadForm = document.getElementById('uploadForm');
const submitBtn = document.getElementById('submitBtn');
const statusDiv = document.getElementById('status');
const resultDiv = document.getElementById('result');
const viewerContainer = document.getElementById('viewerContainer');
const modelList = document.getElementById('modelList');

let scene, camera, renderer, controls;
let currentModel = null;

// Initialize Three.js viewer
function initViewer() {
    const container = document.getElementById('modelViewer');
    const width = container.clientWidth;
    const height = container.clientHeight;

    scene = new THREE.Scene();
    scene.background = new THREE.Color(0xf0f0f0);

    camera = new THREE.PerspectiveCamera(75, width / height, 0.1, 1000);
    camera.position.set(0, 1, 3);

    renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(width, height);
    renderer.shadowMap.enabled = true;
    container.appendChild(renderer.domElement);

    controls = new THREE.OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;

    // Add lights
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
    scene.add(ambientLight);

    const directionalLight = new THREE.DirectionalLight(0xffffff, 0.8);
    directionalLight.position.set(1, 1, 1);
    directionalLight.castShadow = true;
    scene.add(directionalLight);

    animate();
}

function animate() {
    requestAnimationFrame(animate);
    controls.update();
    renderer.render(scene, camera);
}

function loadModel(url) {
    if (currentModel) {
        scene.remove(currentModel);
    }

    const loader = new THREE.GLTFLoader();
    loader.load(url, (gltf) => {
        currentModel = gltf.scene;
        scene.add(currentModel);

        // Center and scale the model
        const box = new THREE.Box3().setFromObject(currentModel);
        const center = box.getCenter(new THREE.Vector3());
        const size = box.getSize(new THREE.Vector3());

        const maxDim = Math.max(size.x, size.y, size.z);
        const scale = 2 / maxDim;
        currentModel.scale.setScalar(scale);

        currentModel.position.sub(center.multiplyScalar(scale));
        currentModel.position.y = -box.min.y * scale;

        viewerContainer.classList.remove('hidden');
    });
}

// Load models list
async function loadModelsList() {
    try {
        const response = await fetch('/api/models');
        
        if (!response.ok) {
            throw new Error('Failed to fetch models: ' + response.status + ' ' + response.statusText);
        }
        
        const models = await response.json();
        
        if (!Array.isArray(models)) {
            throw new Error('Invalid response format');
        }
        
        if (models.length === 0) {
            modelList.innerHTML = '<p>No models generated yet</p>';
        } else {
            modelList.innerHTML = models.map(model => 
                '<div class="model-item" data-id="' + model.id + '">' +
                    '<div class="model-name">' + model.id + '</div>' +
                    '<div class="model-date">' + new Date(model.created_at).toLocaleString() + '</div>' +
                '</div>'
            ).join('');

            // Add click handlers
            modelList.querySelectorAll('.model-item').forEach(item => {
                item.addEventListener('click', () => {
                    document.querySelectorAll('.model-item').forEach(i => i.classList.remove('active'));
                    item.classList.add('active');
                    const modelId = item.dataset.id;
                    loadModel('/download/' + modelId);
                });
            });
        }
    } catch (error) {
        console.error('Error loading models:', error);
        modelList.innerHTML = '<p>Error loading models: ' + error.message + '</p>';
    }
}

// Initialize viewer on page load
window.addEventListener('load', () => {
    console.log('Page loaded, initializing...');
    initViewer();
    console.log('Loading models list...');
    loadModelsList();
});

// Fallback in case load event already fired
if (document.readyState === 'complete') {
    console.log('Document already loaded, initializing immediately...');
    initViewer();
    loadModelsList();
}

uploadArea.addEventListener('click', () => imageInput.click());

uploadArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadArea.classList.add('dragover');
});

uploadArea.addEventListener('dragleave', () => {
    uploadArea.classList.remove('dragover');
});

uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadArea.classList.remove('dragover');
    
    const files = e.dataTransfer.files;
    if (files.length > 0) {
        imageInput.files = files;
        updateUploadArea(files[0].name);
    }
});

imageInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
        updateUploadArea(e.target.files[0].name);
    }
});

function updateUploadArea(filename) {
    uploadArea.innerHTML = '<p>Selected: ' + filename + '</p>';
}

uploadForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = new FormData();
    formData.append('image', imageInput.files[0]);
    
    submitBtn.disabled = true;
    statusDiv.className = 'status processing';
    statusDiv.textContent = 'Uploading and generating 3D model...';
    statusDiv.classList.remove('hidden');
    resultDiv.classList.add('hidden');
    
    try {
        const response = await fetch('/upload', {
            method: 'POST',
            body: formData
        });
        
        if (!response.ok) {
            throw new Error('Upload failed');
        }
        
        const data = await response.json();
        checkStatus(data.job_id);
        
    } catch (error) {
        statusDiv.className = 'status error';
        statusDiv.textContent = 'Error: ' + error.message;
        submitBtn.disabled = false;
    }
});

async function checkStatus(jobId) {
    const interval = setInterval(async () => {
        try {
            const response = await fetch('/status/' + jobId);
            const data = await response.json();
            
            if (data.status === 'completed') {
                clearInterval(interval);
                statusDiv.className = 'status';
                statusDiv.textContent = '3D model generated successfully!';
                resultDiv.innerHTML = '<a href="' + data.model_url + '" download>Download 3D Model (.glb)</a>';
                resultDiv.classList.remove('hidden');
                submitBtn.disabled = false;
                
                // Load the new model in viewer
                loadModel(data.model_url);
                
                // Refresh models list
                loadModelsList();
            } else if (data.status === 'failed') {
                clearInterval(interval);
                statusDiv.className = 'status error';
                statusDiv.textContent = 'Generation failed: ' + (data.error || 'Unknown error');
                submitBtn.disabled = false;
            }
        } catch (error) {
            clearInterval(interval);
            statusDiv.className = 'status error';
            statusDiv.textContent = 'Error checking status';
            submitBtn.disabled = false;
        }
    }, 2000);
}

// Handle window resize
window.addEventListener('resize', () => {
    if (camera && renderer) {
        const container = document.getElementById('modelViewer');
        camera.aspect = container.clientWidth / container.clientHeight;
        camera.updateProjectionMatrix();
        renderer.setSize(container.clientWidth, container.clientHeight);
    }
});
`
