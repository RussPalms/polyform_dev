package modeling_test

// func Test_CollapseTri_ValidMesh(t *testing.T) {
// 	// ARRANGE ================================================================
// 	objString := `# cube.obj
// #

// g cube

// v  0.0  0.0  0.0
// v  0.0  0.0  1.0
// v  0.0  1.0  0.0
// v  0.0  1.0  1.0
// v  1.0  0.0  0.0
// v  1.0  0.0  1.0
// v  1.0  1.0  0.0
// v  1.0  1.0  1.0

// vn  0.0  0.0  1.0
// vn  0.0  0.0 -1.0
// vn  0.0  1.0  0.0
// vn  0.0 -1.0  0.0
// vn  1.0  0.0  0.0
// vn -1.0  0.0  0.0

// f  1//2  7//2  5//2
// f  1//2  3//2  7//2
// f  1//6  4//6  3//6
// f  1//6  2//6  4//6
// f  3//3  8//3  7//3
// f  3//3  4//3  8//3
// f  5//5  7//5  8//5
// f  5//5  8//5  6//5
// f  1//4  5//4  6//4
// f  1//4  6//4  2//4
// f  2//1  6//1  8//1
// f  2//1  8//1  4//1
// `

// 	// ACT ====================================================================
// 	square, err := mesh.FromObj(strings.NewReader(objString))
// 	cm := mesh.NewCollapsableMesh(*square)
// 	cm.CollapseTri(0)
// 	finalMesh := cm.ToMesh()

// 	// ASSERT =================================================================
// 	assert.NoError(t, err)
// 	assert.Equal(t, 10, finalMesh.PrimitiveCount())
// }
