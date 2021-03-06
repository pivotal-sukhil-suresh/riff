/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	v1alpha1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	scheme "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// LinksGetter has a method to return a LinkInterface.
// A group's client should implement this interface.
type LinksGetter interface {
	Links(namespace string) LinkInterface
}

// LinkInterface has methods to work with Link resources.
type LinkInterface interface {
	Create(*v1alpha1.Link) (*v1alpha1.Link, error)
	Update(*v1alpha1.Link) (*v1alpha1.Link, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Link, error)
	List(opts v1.ListOptions) (*v1alpha1.LinkList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Link, err error)
	LinkExpansion
}

// links implements LinkInterface
type links struct {
	client rest.Interface
	ns     string
}

// newLinks returns a Links
func newLinks(c *ProjectriffV1alpha1Client, namespace string) *links {
	return &links{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the link, and returns the corresponding link object, and an error if there is any.
func (c *links) Get(name string, options v1.GetOptions) (result *v1alpha1.Link, err error) {
	result = &v1alpha1.Link{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("links").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Links that match those selectors.
func (c *links) List(opts v1.ListOptions) (result *v1alpha1.LinkList, err error) {
	result = &v1alpha1.LinkList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("links").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested links.
func (c *links) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("links").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a link and creates it.  Returns the server's representation of the link, and an error, if there is any.
func (c *links) Create(link *v1alpha1.Link) (result *v1alpha1.Link, err error) {
	result = &v1alpha1.Link{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("links").
		Body(link).
		Do().
		Into(result)
	return
}

// Update takes the representation of a link and updates it. Returns the server's representation of the link, and an error, if there is any.
func (c *links) Update(link *v1alpha1.Link) (result *v1alpha1.Link, err error) {
	result = &v1alpha1.Link{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("links").
		Name(link.Name).
		Body(link).
		Do().
		Into(result)
	return
}

// Delete takes name of the link and deletes it. Returns an error if one occurs.
func (c *links) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("links").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *links) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("links").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched link.
func (c *links) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Link, err error) {
	result = &v1alpha1.Link{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("links").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
